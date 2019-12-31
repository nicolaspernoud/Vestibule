export function AnimateCSS(el, animationName, callback) {
  el.classList.add("animated", "faster", animationName);
  function handleAnimationEnd() {
    el.classList.remove("animated", "faster", animationName);
    el.removeEventListener("animationend", handleAnimationEnd);
    if (typeof callback === "function") callback();
  }
  el.addEventListener("animationend", handleAnimationEnd);
}
