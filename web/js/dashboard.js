var player;
var Username;
var AppPermissionsLevel;
var ChannelName;
var ChannelPermissionsLevel;
var formObj;
var activeMessages = 0;
var Channels;;

const PermsAfterClass = "displayPerms"
const PermsStringAttr = "permName"
const ChannelCookieName = "selectedChannel";
const UserCookieName = "User"
const ChannelDropdownID = "channelDropdown";
const ChannelDropdownPopoverID = "channelList";
const TitleFormID = "streamTitleForm";
const TitleFormInputFieldID = "streamTitle";
const UserMenuID = "userMenu";
const PermsDropdownID = "permsDropdown";
const PermsDropdownPopoverID = "permsList";
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Set stream title
document.getElementById(TitleFormID).addEventListener("submit", async (e) => {
  e.preventDefault();
  const formData = new FormData();
  const URL = "/twitch/setStreamTitle";
  
  formData.append("channel", ChannelName)
  formData.append("title", document.getElementById(TitleFormInputFieldID).value)
  try {
    fetch(URL, {
      method: "POST",
      body: formData
    })
    .then(response => response.text())
    .then(data => {
      displayMessage(data, false)
    })
  } catch (e) {
    console.error(e);
    displayMessage(e, true)
  }
});

async function init() {
  try {
    ChannelName = getCookie(ChannelCookieName).split(':')[0];
    ChannelPermissionsLevel = getCookie(ChannelCookieName).split(':')[1];
  } catch {
    ChannelName = null
    ChannelPermissionsLevel = null
  }

  Username = getCookie("User").split(':')[0];
  AppPermissionsLevel = getCookie("User").split(':')[1];
  populateChannelDropdown(ChannelDropdownID, ChannelDropdownPopoverID);
  populatePermissionsDropdown(PermsDropdownID, PermsDropdownPopoverID, false);
  getStreamTitle(TitleFormInputFieldID);
  loadChannelEmbed(ChannelName);
  showUsername(UserMenuID);
}

async function rotateOnPress(e) {
  if (e.classList.contains("rotate")) {
    return;
  }
  e.classList.add("rotate");
  await delay(1000);
  e.classList.remove("rotate");
}

function changeChannel(e) {
  const dropdownBtn = document.getElementById(ChannelDropdownID);
  const passedChannelName = e.innerText;
  const passedChannelPerms = e.value;
  const cookieValue = passedChannelName+":"+passedChannelPerms;

  ChannelName = passedChannelName;
  ChannelPermissionsLevel = passedChannelPerms;

  dropdownBtn.setAttribute(PermsStringAttr, getPermName(ChannelPermissionsLevel));
  dropdownBtn.classList.add(PermsAfterClass);
  dropdownBtn.innerText = passedChannelName;
  document.getElementById(ChannelDropdownPopoverID).hidePopover();
  
  getStreamTitle(TitleFormInputFieldID);
  player.setChannel(passedChannelName);
  setCookie(ChannelCookieName, cookieValue, 365);
}

async function populateChannelDropdown(dropdownID, dropdownPopoverID) {
  const URL = "/user/GetUserChannels";
  const ul = document.getElementById(ChannelDropdownPopoverID)
  var selectElement = document.getElementById(dropdownID);
  try {
    const request = new Request(URL);
    const response = await fetch(request);
    const text = await response.text();
    const parsedJSON = JSON.parse(text);
    Channels = parsedJSON;

    selectElement.innerHTML = '';
    ul.textContent = ""
    Object.keys(parsedJSON).forEach(key => {
      const newLi = document.createElement("li");
      let text = document.createTextNode(key);
      newLi.setAttribute(PermsStringAttr, getPermName(parsedJSON[key]));
      newLi.classList.add(PermsAfterClass)
      newLi.appendChild(text);
      newLi.value = parsedJSON[key]
      newLi.after
      newLi.onclick = function() { changeChannel(this); }
      ul.appendChild(newLi);
    });


    // If no cookie is present, select the first channel
    if (ChannelName === null) {
      selectElement.value = ul.children[0].value;
      selectElement.innerText = ul.children[0].innerText;
      changeChannel(selectElement);
    } else {
      selectElement.innerText = ChannelName;
      selectElement.value = ChannelName;
      selectElement.classList.add(PermsAfterClass);
      selectElement.setAttribute(PermsStringAttr, getPermName(ChannelPermissionsLevel));
    }
  } catch (e) {
    console.error(e);
    displayMessage(e, true)
  }
}

function loadChannelEmbed(channelName) {
  if (channelName === null) { channelName = "Twitch"}
  if (player != null) { document.getElementById("twitch-embed").innerHTML = ''; }
  player = new Twitch.Player("twitch-embed", {
    channel: channelName,
    width: "100%",
    height: "100%",
    autoplay: false,
    muted: true,
    time: "0h0m0s"
  });
}

function getStreamTitle(fieldName) {
  const e = document.getElementById(fieldName)
  const URL = "/twitch/getStreamTitle";
  try {
    fetch(URL+"?channel="+ChannelName)
    .then(response => response.text())
    .then(data => {
      e.value = data
      console.log(e)
    })
  } catch (e) {
    console.error("getStreamTitleError");
    console.error(e);
    displayMessage(e, true)
  }
}

function showUsername(fieldName) {
  const e = document.getElementById(fieldName);
  const label = e.getElementsByTagName("label")[0]
  label.classList.add(PermsAfterClass)
  label.setAttribute(PermsStringAttr, AppPermissionsLevel)
  label.innerText = Username;
}

function changePassword(e) {
  if (!(e.newPassword.value == e.newPasswordRepeat.value)) {
    displayMessage("Passwords don't match!", true)
    return false
  } else {
    const formData = new FormData();
    const URL = "/user/changePassword";
    
    formData.append("user", Username)
    formData.append("pass", e.oldPassword.value)
    formData.append("newPass", e.newPassword.value)
    try {
      fetch(URL, {
        method: "POST",
        body: formData
      })
      .then(response => {
        if (response.ok) {
          response.text().then( text => {
            displayMessage(text, false)
            e.reset()
            e.parentNode.close()
          })
        } else {
          response.text().then( text => {
            displayMessage(text, true)
          })
        }
      })
    } catch (e) {
      console.error(e);
      displayMessage(e, true)
    }
  }
}

// dropdownID string - ul element that the il elements will be pushed to
// dropdownPopoverID string - element that opens the dropdown and shows the selected value
// channelPerms bool - false is for general App permissions, true is for channels
function populatePermissionsDropdown(dropdownID, dropdownPopoverID, channelPerms) {
  const ul = document.getElementById(dropdownPopoverID);
  const selectElement = document.getElementById(dropdownID);
  ul.innerText = '';
  Object.keys(userPerms).forEach(key => {
    if (userPerms[key] < AppPermissionsLevel) {
      const newLi = document.createElement("li");
      let text = document.createTextNode(key);
      newLi.appendChild(text);
      newLi.value = userPerms[key]
      newLi.onclick = function() {
        selectElement.innerText = newLi.innerText;
        ul.hidePopover();
      }
      ul.appendChild(newLi);
    }
  });
}

function registerNewUser(e) {
  const user = e.user.value;
  const pass = e.pass.value;
  const passRepeat = e.passRepeat.value;
  console.log(user)
  console.log(pass)
  console.log(passRepeat)
  if (!(pass == passRepeat)) {
    displayMessage("Passwords don't match!", true)
    return
  } else {
    displayMessage("Passwords match!", false)
    return
  }
}