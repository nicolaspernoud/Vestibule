// Imports
import * as Messages from "/services/messages/messages.js";
import * as Auth from "/services/auth/auth.js";
import { AnimateCSS } from "/services/common/common.js";

export class Open {
  constructor(hostname, readwrite, files, file) {
    this.hostname = hostname;
    this.readonly = !readwrite;
    this.files = files;
    this.file = file;
    this.index = files.findIndex(element => element.name === file.name);
    this.type = GetType(this.file);
    this.url = `${hostname}${file.path}`;
  }

  update(isNext) {
    const idx = isNext ? this.index + 1 : this.index - 1;
    if (idx >= 0 && idx < this.files.length) {
      this.index = idx;
      this.file = this.files[idx];
      this.type = GetType(this.file);
      this.url = `${this.hostname}${this.file.path}`;
      if (this.type && !this.file.isDir) {
        this.openModal.parentNode.removeChild(this.openModal);
        this.show(false);
      }
    }
  }

  async show(animated) {
    this.openModal = document.createElement("div");
    this.openModal.classList.add("modal", "is-active");
    if (animated) this.openModal.classList.add("animated", "fadeIn", "faster");
    let content;
    if (this.type == "text") {
      try {
        const response = await fetch(this.url, {
          method: "get"
        });
        if (response.status !== 200) {
          throw new Error(`Text content could not be fetched (status ${response.status})`);
        }
        content = await response.text();
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
    }
    this.openModal.innerHTML = this.computeTemplate(content);
    this.openModal.querySelector("#" + "open-close").addEventListener("click", () => {
      AnimateCSS(this.openModal, "fadeOut", () => {
        this.openModal.parentNode.removeChild(this.openModal);
      });
    });
    this.openModal.querySelector("#" + "open-previous").addEventListener("click", () => {
      this.update(false);
    });
    this.openModal.querySelector("#" + "open-next").addEventListener("click", () => {
      this.update(true);
    });
    document.body.appendChild(this.openModal);
  }

  computeTemplate(content) {
    return /* HTML */ `
      <div class="modal-content">
        <div class="box">
          ${
            this.type == "other"
              ? /* HTML */ `
                  <object data="${this.url}"></object>
                `
              : ""
          }
          ${this.type == "image" ? `<img src="${this.url}" alt="Previewed image" />` : ""}
          ${
            this.type == "audio"
              ? /* HTML */ `
                  <audio controls autoplay><source src="${this.url}" /></audio>
                `
              : ""
          }
          ${
            this.type == "video"
              ? /* HTML */ `
                  <video controls autoplay><source src="${this.url}" /></video>
                `
              : ""
          }
          ${
            this.type == "text"
              ? /* HTML */ `
                  <textarea class="textarea" readonly>${content}</textarea>
                `
              : ""
          }
          <h1>${this.file.name}</h1>
          </br>
          <div class="buttons">
            <button id="open-previous" class="button">
              <span class="icon is-small"><i class="fas fa-arrow-circle-left"></i></span>
            </button>
            <button id="open-next" class="button">
              <span class="icon is-small"><i class="fas fa-arrow-circle-right"></i></span>
            </button>
            <button id="open-share" class="button" disabled>
              <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
            </button>
            <button id="open-close" class="button">
              <span class="icon is-small"><i class="fas fa-times"></i></span>
            </button>
          </div>
        </div>
      </div>
    `;
  }
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
