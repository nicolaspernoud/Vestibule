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

export function GetType(file) {
  if (/(txt|md|csv|sh|nfo|log|json|yml|srt)$/.test(file.name.toLowerCase())) {
    return "text";
  }
  if (/(docx|doc|odt|xlsx|xls|ods|pptx|ppt|opd)$/.test(file.name.toLowerCase())) {
    return "document";
  }
  if (/(jpg|png|gif|svg|jpeg)$/.test(file.name.toLowerCase())) {
    return "image";
  }
  if (/(mp3|wav|ogg)$/.test(file.name.toLowerCase())) {
    return "audio";
  }
  if (/(mp4|avi|mkv|m4v)$/.test(file.name.toLowerCase())) {
    return "video";
  }
  if (/(pdf)$/.test(file.name.toLowerCase())) {
    return "other";
  }
}
