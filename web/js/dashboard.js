var player;
var SelectedChannel
const ChannelDropdownID = "channelDropdown"
const NameFormID = "channelNameForm"
const NameFormInputFieldID = "channelName"
const TitleFormID = "streamTitleForm"
const TitleFormInputFieldID = "streamTitle"
const ChannelCookieName = "selectedChannel"
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
document.getElementById(NameFormID).addEventListener("submit", function (e) {
  e.preventDefault();
  getChannelID(e.target)
});

document.getElementById(TitleFormID).addEventListener("submit", function (e) {
  e.preventDefault();
  setStreamTitle(e)
});

async function init() {
  SelectedChannel = getCookie(ChannelCookieName);
  if (SelectedChannel == null) {
    SelectedChannel = "0" // should let the twitch player load, but fail to find a video if no cookie is set
  }
  console.log(SelectedChannel);
  populateChannelDropdown(ChannelDropdownID);
  loadChannelEmbed(SelectedChannel);
  getStreamTitle(TitleFormInputFieldID);
}

async function rotateOnPress(e) {
  if (e.classList.contains("rotate")) {
    return;
  }
  e.classList.add("rotate");
  await delay(1000);
  e.classList.remove("rotate");
}

async function getChannelID(form) {
  const URL = "/twitch/getChannelID";
  var formData = new FormData(form);
  try {
    fetch(URL, {
      method: "POST",
      body: formData
    })
    .then(response => response.text())
    .then(data => {
      document.getElementById(NameFormInputFieldID).value = ""
        document.getElementById(NameFormInputFieldID).placeholder = data;
    })
  } catch (e) {
    console.error(e);
    document.getElementById(NameFormInputFieldID).value = e;
  }
}

function changeChannel(e) {
  player.setChannel(e.value);
  setCookie(ChannelCookieName, e.value, 365);
}
async function populateChannelDropdown(dropdownID) {
  const URL = "/user/GetUserChannels";
  var selectElement = document.getElementById(dropdownID);
  try {
    const request = new Request(URL);
    const response = await fetch(request);
    const text = await response.text();
    const parsedJSON = JSON.parse(text);
    selectElement.innerHTML = '';
    Object.keys(parsedJSON).forEach(key => {
      selectElement.appendChild(new Option(key, key)).cloneNode(true)
    });
    // If no cookie is present, select the first channel
    if (SelectedChannel == 0) {
      selectElement.value = selectElement[0].value
    } else {
      selectElement.value = SelectedChannel
    }
  } catch (e) {
    console.error(e);
    document.getElementById(NameFormInputFieldID).value = e;
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
  e.value = "PH - " + SelectedChannel + " - PH";
}
function setStreamTitle(e) {
  console.log(e.value)
}
