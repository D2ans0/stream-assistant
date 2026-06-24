const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
document.getElementById("channelNameForm").addEventListener("submit", function (e) {
  e.preventDefault();
  getChannelID(e.target)
});
function init() {
  populateStreamTitleChannelDropdown()
}
async function rotateOnPress(e) {
  if (e.classList.contains("rotate")) {
    return
  }
  e.classList.add("rotate")
  await delay(1000)
  e.classList.remove("rotate")
}
function getChannelID(form) {
  const URL = "/twitch/getChannelID";
  var formData = new FormData(form);
  try {
    fetch(URL, {
      method: "POST",
      body: formData
    })
    .then(response => response.text())
    .then(data => {
        document.getElementById("channelNameFormOutput").innerHTML = data;
    })
  } catch (e) {
    console.error(e);
    document.getElementById("channelNameFormOutput").innerHTML = e;
  }
}
function sendChannelStreamTitle() {
  console.log("changed title")
}
function populateStreamTitleChannelDropdown() {
  // placeholder, need to add api call for getting a user's accessible channels
  let data = [
      "name_1",
      "name_2"
  ];
  var selectElement = document.getElementById('streamTitleChannelDropdown');
  console.log(selectElement.childNodes)
  selectElement.innerHTML = '';
  data.map(item => selectElement.appendChild(new Option(item, item)).cloneNode(true));
}