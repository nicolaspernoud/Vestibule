// Imports
import * as Messages from "/services/messages/messages.js";
import { AnimateCSS, RandomString, GetType, GID } from "/services/common/common.js";
import { Share } from "/components/davs/share.js";
import * as Auth from "/services/auth/auth.js";
import { LoadImage } from "/components/davs/explorer.js";

export class Open {
  constructor(hostname, files, file) {
    this.hostname = hostname;
    this.files = files;
    this.file = file;
    this.index = files.findIndex(element => element.name === file.name);
    this.type = GetType(this.file);
    this.url = `${hostname}${file.path}`;
    // Random id seed
    this.prefix = RandomString(8);
  }

  gid(id) {
    return GID(this, id);
  }

  update(isNext) {
    const idx = isNext ? this.index + 1 : this.index - 1;
    if (idx >= 0 && idx < this.files.length) {
      const type = GetType(this.files[idx]);
      if (type && type !== "document" && !this.files[idx].isDir) {
        this.index = idx;
        this.file = this.files[idx];
        this.type = type;
        this.url = `${this.hostname}${this.file.path}`;
        this.openModal.parentNode.removeChild(this.openModal);
        this.show(false);
      }
    }
  }

  async show(animated) {
    this.user = await Auth.GetUser();
    this.openModal = document.createElement("div");
    this.openModal.classList.add("modal", "is-active");
    if (animated) this.openModal.classList.add("animated", "fadeIn", "faster");
    let content;
    if (this.type == "text") {
      try {
        const response = await fetch(this.url, {
          method: "get",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include"
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
    document.body.appendChild(this.openModal);
    if (this.type === "image") {
      LoadImage(this.gid("open-image"), this.url, this.user);
    }
    this.gid("open-close").addEventListener("click", () => {
      AnimateCSS(this.openModal, "fadeOut", () => {
        this.openModal.parentNode.removeChild(this.openModal);
      });
    });
    this.gid("open-previous").addEventListener("click", () => {
      this.update(false);
    });
    this.gid("open-next").addEventListener("click", () => {
      this.update(true);
    });
    this.gid("open-share").addEventListener("click", () => {
      const shareModal = new Share(this.hostname, this.file);
      shareModal.show();
    });
  }

  computeTemplate(content) {
    return /* HTML */ `
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">${this.file.name}</p>
          <button class="delete" aria-label="close" id="${this.prefix}open-close"></button>
        </header>
        <div class="modal-content">
          <div class="box">
            ${this.type == "other"
              ? /* HTML */ `
                  <object data="${this.url}"></object>
                `
              : ""}
            ${this.type == "image" ? `<img id="${this.prefix}open-image" src="assets/spinner.svg" alt="Previewed image" />` : ""}
            ${this.type == "audio"
              ? /* HTML */ `
                  <audio controls autoplay><source src="${this.url}" /></audio>
                `
              : ""}
            ${this.type == "video"
              ? /* HTML */ `
                  <video controls autoplay><source src="${this.url}" /></video>
                `
              : ""}
            ${this.type == "text"
              ? /* HTML */ `
                  <textarea class="textarea" readonly>${content}</textarea>
                `
              : ""}
            <br />
            <div class="buttons">
              <button id="${this.prefix}open-previous" class="button">
                <span class="icon is-small"><i class="fas fa-arrow-circle-left"></i></span>
              </button>
              <button id="${this.prefix}open-next" class="button">
                <span class="icon is-small"><i class="fas fa-arrow-circle-right"></i></span>
              </button>
              <button id="${this.prefix}open-share" class="button">
                <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
  }
}
