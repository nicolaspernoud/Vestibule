// Imports
import { AnimateCSS, RandomString, GetType, GID } from "/services/common/common.js";
import { Share } from "/components/davs/share.js";
import { HandleError } from "/services/common/errors.js";

export class Open {
  constructor(user, hostname, fullHostname, files, file) {
    this.user = user;
    this.hostname = hostname;
    this.fullHostname = fullHostname;
    this.files = files;
    this.file = file;
    this.index = files.findIndex((element) => element.name === file.name);
    this.type = GetType(this.file);
    this.url = `${fullHostname}${file.path}`;
    // Random id seed
    this.prefix = RandomString(8);
  }

  gid(id) {
    return GID(this, id);
  }

  async show(animated) {
    // Create empty modal
    this.openModal = document.createElement("div");
    this.openModal.classList.add("modal", "is-active");
    if (animated) this.openModal.classList.add("animate__animated", "animate__fadeIn");
    this.openModal.innerHTML = this.emptyTemplate();
    document.body.appendChild(this.openModal);
    // Add events
    this.gid("open-close").addEventListener("click", async () => {
      await AnimateCSS(this.openModal, "fadeOut");
      this.openModal.parentNode.removeChild(this.openModal);
    });
    this.gid("open-previous").addEventListener("click", () => {
      this.seek(false);
    });
    this.gid("open-next").addEventListener("click", () => {
      this.seek(true);
    });
    this.gid("open-share").addEventListener("click", () => {
      const shareModal = new Share(this.user, this.hostname, this.file);
      shareModal.show();
    });
    // Display
    this.showContent();
  }

  seek(isNext) {
    const idx = isNext ? this.index + 1 : this.index - 1;
    if (idx >= 0 && idx < this.files.length) {
      const type = GetType(this.files[idx]);
      if (type && type !== "document" && !this.files[idx].isDir) {
        this.index = idx;
        this.file = this.files[idx];
        this.type = type;
        this.url = `${this.fullHostname}${this.file.path}`;
        this.showContent();
      }
    }
  }

  async showContent() {
    let content;
    let token;
    if (this.type == "text") {
      try {
        const response = await fetch(this.url, {
          method: "GET",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken,
          }),
          credentials: "include",
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
            "XSRF-Token": this.user.xsrftoken,
          }),
          credentials: "include",
          body: JSON.stringify({
            sharedfor: "file_preview",
            lifespan: 1,
            url: this.hostname + this.file.path,
            readonly: true,
          }),
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
    this.updateTemplate(content, token);
  }

  emptyTemplate() {
    return /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <button class="delete navbar-menu-icon" aria-label="close" id="${this.prefix}open-close"></button>
          <p id="${this.prefix}filename" class="modal-card-title has-text-centered">${this.file.name}</p>
        </header>
        <section id="${this.prefix}content" class="modal-card-body is-paddingless flex-container"></section>
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

  updateTemplate(content, token) {
    this.gid("content").innerHTML = /* HTML */ `
      ${this.type == "other" ? /* HTML */ ` <embed src="${this.url}?token=${token}&inline" type="application/pdf" width="100%" style="height: 75vh;" /> ` : ""}
      ${this.type == "image" ? `<img id="${this.prefix}open-image" src="${this.url}?token=${token}" alt="Previewed image" />` : ""}
      ${this.type == "audio" ? /* HTML */ ` <audio controls autoplay><source src="${this.url}?token=${token}" /></audio> ` : ""}
      ${this.type == "video" ? /* HTML */ ` <video controls autoplay><source src="${this.url}?token=${token}" /></video> ` : ""}
      ${this.type == "text" ? /* HTML */ ` <textarea class="textarea" readonly>${content}</textarea> ` : ""}
    `;
    this.gid("filename").innerHTML = /* HTML */ `${this.file.name}`;
  }
}
