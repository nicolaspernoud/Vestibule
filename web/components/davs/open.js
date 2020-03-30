// Imports
import { AnimateCSS, RandomString, GetType, GID } from "/services/common/common.js";
import { Share } from "/components/davs/share.js";
import * as Auth from "/services/auth/auth.js";
import { HandleError } from "/services/common/errors.js";

export class Open {
  constructor(hostname, fullHostname, files, file) {
    this.hostname = hostname;
    this.fullHostname = fullHostname;
    this.files = files;
    this.file = file;
    this.index = files.findIndex(element => element.name === file.name);
    this.type = GetType(this.file);
    this.url = `${fullHostname}${file.path}`;
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
        this.url = `${this.fullHostname}${this.file.path}`;
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
    let token;
    if (this.type == "text") {
      try {
        const response = await fetch(this.url, {
          method: "GET",
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
        HandleError(e);
      }
    } else {
      try {
        const response = await fetch(location.origin + "/api/common/Share", {
          method: "POST",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include",
          body: JSON.stringify({
            sharedfor: "file_preview",
            lifespan: 1,
            url: this.hostname + this.file.path,
            readonly: true
          })
        });
        if (response.status !== 200) {
          throw new Error(`Share token could not be made (status ${response.status})`);
        }
        token = await response.text();
        token = encodeURIComponent(token);
      } catch (e) {
        HandleError(e);
      }
    }
    this.openModal.innerHTML = this.computeTemplate(content, token);
    document.body.appendChild(this.openModal);
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

  computeTemplate(content, token) {
    return /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <button class="delete navbar-menu-icon" aria-label="close" id="${this.prefix}open-close"></button>
          <p class="modal-card-title has-text-centered">${this.file.name}</p>
        </header>
        <section class="modal-card-body is-paddingless flex-container">
          ${this.type == "other"
            ? /* HTML */ `
                <embed src="${this.url}?token=${token}&inline" type="application/pdf" width="100%" style="height: 75vh;" />
              `
            : ""}
          ${this.type == "image" ? `<img id="${this.prefix}open-image" src="${this.url}?token=${token}" alt="Previewed image" />` : ""}
          ${this.type == "audio"
            ? /* HTML */ `
                <audio controls autoplay><source src="${this.url}?token=${token}" /></audio>
              `
            : ""}
          ${this.type == "video"
            ? /* HTML */ `
                <video controls autoplay><source src="${this.url}?token=${token}" /></video>
              `
            : ""}
          ${this.type == "text"
            ? /* HTML */ `
                <textarea class="textarea" readonly>${content}</textarea>
              `
            : ""}
        </section>
        <footer class="modal-card-foot">
          <button id="${this.prefix}open-previous" class="button">
            <span class="icon is-small"><i class="fas fa-arrow-circle-left"></i></span>
          </button>
          <button id="${this.prefix}open-next" class="button">
            <span class="icon is-small"><i class="fas fa-arrow-circle-right"></i></span>
          </button>
          <button id="${this.prefix}open-share" class="button">
            <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
          </button>
        </footer>
      </div>
    `;
  }
}
