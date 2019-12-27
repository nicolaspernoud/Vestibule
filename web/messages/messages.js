export function Show(bulmaClass, message) {
  let msg = document.createElement("div");
  msg.innerText = message;
  msg.classList.add("notification", "animated", "fadeInUp", "faster");
  msg.classList.add(bulmaClass);
  const delBtn = document.createElement("button");
  delBtn.classList.add("delete");
  msg.appendChild(delBtn);
  document.body.appendChild(msg);
  const timer = setTimeout(function() {
    msg.parentNode.removeChild(msg);
  }, 5000);
  delBtn.addEventListener("click", function() {
    msg.parentNode.removeChild(msg);
    clearTimeout(timer);
  });
}
