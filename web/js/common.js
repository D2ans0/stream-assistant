const messageListID = "messageList"
const errorMessageClass = "errorMessage"
const userPerms = {"User": 1, "Moderator": 2, "Admin": 3, "Owner": 4}

// cookie functions taken from https://stackoverflow.com/a/24103596
function setCookie(name,value,days) {
    var expires = "";
    if (days) {
        var date = new Date();
        date.setTime(date.getTime() + (days*24*60*60*1000));
        expires = "; expires=" + date.toUTCString();
    }
    document.cookie = name + "=" + (value || "")  + expires + "; path=/";
}
function getCookie(name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for(var i=0;i < ca.length;i++) {
        var c = ca[i];
        while (c.charAt(0)==' ') c = c.substring(1,c.length);
        if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
    }
    return null;
}
function eraseCookie(name) {   
    document.cookie = name +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

async function displayMessage(message, isError) {
  const container = document.getElementById(messageListID)
  container.showPopover();
  activeMessages += 1;
  let e = document.createElement("div");
  let text = document.createTextNode(message);
  e.appendChild(text);
  if (isError) {
    e.classList.add(errorMessageClass);
  }
  document.getElementById(messageListID).prepend(e);
  await delay(4000 + 1000*activeMessages); // add delay if there's messages already
  e.style.transform = "translateY(-1000px)";
  await delay(1000);
  e.remove();
  activeMessages -= 1;
  if (activeMessages == 0) {
    container.hidePopover()
  }
}

function getPermName(permissionLevel) {
    permissionLevel = Number(permissionLevel)
    return Object.keys(userPerms).find(val=> userPerms[val] === permissionLevel);
}