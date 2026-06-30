var player;
var SelectedChannel;
var Username;
var PermissionsLevel;
var formObj;
var activeMessages = 0;

const ChannelCookieName = "selectedChannel";
const UserCookieName = "User"
const ChannelDropdownID = "channelDropdown";
const ChannelDropdownPopoverID = "channelList";
const TitleFormID = "streamTitleForm";
const TitleFormInputFieldID = "streamTitle";
const UserMenuID = "userMenu";
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Set stream title
document.getElementById(TitleFormID).addEventListener("submit", async (e) => {
  e.preventDefault();
  const formData = new FormData();
  const URL = "/twitch/setStreamTitle";
  
  formData.append("channel", SelectedChannel)
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
  console.log(e.value)

});

async function init() {
  SelectedChannel = getCookie(ChannelCookieName);
  if (SelectedChannel == null) {
    SelectedChannel = "0" // should let the twitch player load, but fail to find a video if no cookie is set
  }
  console.log(SelectedChannel);
  populateChannelDropdown(ChannelDropdownID);
  getStreamTitle(TitleFormInputFieldID);
  loadChannelEmbed(SelectedChannel);
  UserName = getCookie("User").split(':')[0];
  PermissionsLevel = getCookie("User").split(':')[1];
  setUsername(UserMenuID);
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
  console.log(e.innerText);
  document.getElementById(ChannelDropdownPopoverID).hidePopover()
  document.getElementById(ChannelDropdownID).innerText = e.innerText
  getStreamTitle(TitleFormInputFieldID);
  player.setChannel(e.innerText);
  setCookie(ChannelCookieName, e.innerText, 365);
}

async function populateChannelDropdown(dropdownID) {
  const URL = "/user/GetUserChannels";
  const ul = document.getElementById(ChannelDropdownPopoverID)
  var selectElement = document.getElementById(dropdownID);
  try {
    const request = new Request(URL);
    const response = await fetch(request);
    const text = await response.text();
    const parsedJSON = JSON.parse(text);
    selectElement.innerHTML = '';
    ul.textContent = ""
    Object.keys(parsedJSON).forEach(key => {
      const newLi = document.createElement("li");
      let text = document.createTextNode(key);
      newLi.appendChild(text);
      newLi.onclick = function() { changeChannel(this) }
      ul.appendChild(newLi);
    });


    // If no cookie is present, select the first channel
    if (SelectedChannel == 0) {
      selectElement.value = ul.children[0].innerText
      selectElement.innerText = ul.children[0].innerText
      changeChannel(selectElement)
    } else {
      selectElement.innerText = SelectedChannel
      selectElement.value = SelectedChannel
    }
  } catch (e) {
    console.error(e);
    displayMessage(e, true)
  }
}

function loadChannelEmbed(channelName) {
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
  e = document.getElementById(fieldName)
  const URL = "/twitch/getStreamTitle";
  try {
    fetch(URL+"?channel="+SelectedChannel)
    .then(response => response.text())
    .then(data => {
      e.value = data
    })
  } catch (e) {
    console.error("getStreamTitleError");
    console.error(e);
    displayMessage(e, true)
  }
}

function setUsername(fieldName) {
  e = document.getElementById(fieldName);
  e.getElementsByTagName("label")[0].innerText = UserName;
}

function changePassword(e) {
  formObj = e;
  console.log("Start")
  if (!(e.newPassword.value == e.newPasswordRepeat.value)) {
    displayMessage("Passwords don't match!", true)
    return false
  } else {
    displayMessage("Passwords match!", false)
    console.log("Passwords match!")
    return true
    const formData = new FormData();
    const URL = "/twitch/setStreamTitle";
    
    formData.append("user", Username)
    formData.append("oldPassword", e.oldPassword.value)
    formData.append("newPassword", e.newPassword.value)
    try {
      fetch(URL, {
        method: "POST",
        body: formData
      })
      .then(response => response.text())
      .then(data => {
        displayMessage(e, false)
      })
    } catch (e) {
      console.error(e);
      displayMessage(e, true)
    }
  }
}

async function displayMessage(message, isError) {
  container = document.getElementById("messageList")
  container.showPopover();
  activeMessages += 1;
  let e = document.createElement("div");
  let text = document.createTextNode(message);
  e.appendChild(text);
  if (isError) {
    e.classList.add("errorMessage");
  }
  document.getElementById("messageList").prepend(e);
  await delay(4000 + 1000*activeMessages); // add delay if there's messages already
  e.style.transform = "translateY(-1000px)";
  await delay(1000);
  e.remove();
  activeMessages -= 1;
  if (activeMessages == 0) {
    container.hidePopover()
  }
}