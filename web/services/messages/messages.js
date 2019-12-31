// Imports
import { AnimateCSS } from "/services/common/common.js";

let offset = 0;
let messages = [];

export function Show(bulmaClass, message) {
  // Remove duplicate messages
  if (!messages.includes(message)) {
    let msg = document.createElement("div");
    msg.style.marginBottom = offset.toString() + "px";
    msg.innerText = message;
    msg.classList.add("notification", "animated", "fadeInUp", "faster");
    msg.classList.add(bulmaClass);
    const delBtn = document.createElement("button");
    delBtn.classList.add("delete");
    msg.appendChild(delBtn);
    document.body.appendChild(msg);
    const height = msg.offsetHeight + 1;
    offset = offset + height;
    messages.push(message);
    const timer = setTimeout(function() {
      removeMsg(msg, message, height);
    }, 5000);
    delBtn.addEventListener("click", function() {
      removeMsg(msg, message, height);
      clearTimeout(timer);
    });
  }
}

function removeMsg(msg, message, height) {
  AnimateCSS(msg, "fadeOutDown", function() {
    msg.parentNode.removeChild(msg);
  });
  offset = offset - height;
  const index = messages.indexOf(message);
  if (index > -1) {
    messages.splice(index, 1);
  }
}
