export function AnimateCSS(el, animationName, callback) {
  el.classList.add("animated", "faster", animationName);
  function handleAnimationEnd() {
    el.classList.remove("animated", "faster", animationName);
    el.removeEventListener("animationend", handleAnimationEnd);
    if (typeof callback === "function") callback();
  }
  el.addEventListener("animationend", handleAnimationEnd);
}

export let GID = (obj, id) => {
  return document.getElementById(obj.prefix + id);
};

export function RandomString(length) {
  let text = "";
  const possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < length; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
