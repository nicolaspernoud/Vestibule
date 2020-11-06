export const AnimateCSS = (element, animation, prefix = "animate__") =>
  // We create a Promise and return it
  new Promise((resolve, reject) => {
    const animationName = `${prefix}${animation}`;

    element.classList.add(`${prefix}animated`, animationName);

    // When the animation ends, we clean the classes and resolve the Promise
    function handleAnimationEnd() {
      element.classList.remove(`${prefix}animated`, animationName);
      element.removeEventListener("animationend", handleAnimationEnd);
      resolve("Animation ended");
    }
    element.addEventListener("animationend", handleAnimationEnd);
  });

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

export const Truncate = (input) => (input.length > 12 ? `${input.substring(0, 12)}...` : input);

export function EncodeURIWithSpecialsCharacters(str) {
  return encodeURI(str).replace(/[!'()*]/g, escape);
}
