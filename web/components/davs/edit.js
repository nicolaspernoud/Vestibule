// Imports
import * as Messages from "/services/messages/messages.js";
import { AnimateCSS, RandomString, GID } from "/services/common/common.js";
import { Share } from "/components/davs/share.js";
import * as Auth from "/services/auth/auth.js";

export class Edit {
  constructor(hostname, file) {
    this.hostname = hostname;
    this.file = file;
    this.url = `${hostname}${file.path}`;
    // Random id seed
    this.prefix = RandomString(8);
  }

  gid(id) {
    return GID(this, id);
  }

  async show() {
    this.user = await Auth.GetUser();
    this.editModal = document.createElement("div");
    this.editModal.classList.add("modal", "is-active");
    this.editModal.classList.add("animated", "fadeIn", "faster");
    let content;
    try {
      const response = await fetch(this.url, {
        method: "get",
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
    this.editModal.innerHTML = this.computeTemplate(content);
    document.body.appendChild(this.editModal);
    this.gid("edit-close").addEventListener("click", () => {
      AnimateCSS(this.editModal, "fadeOut", () => {
        this.editModal.parentNode.removeChild(this.editModal);
      });
    });
    this.gid("edit-save").addEventListener("click", () => {
      this.save();
    });
    this.gid("edit-share").addEventListener("click", () => {
      const shareModal = new Share(this.hostname, this.file);
      shareModal.show();
    });
  }

  computeTemplate(content) {
    return /* HTML */ `
    <div class="modal-content">
        <div class="box">
          <textarea id="${this.prefix}edit-content" class="textarea">${content}</textarea>
          <h1>${this.file.name}</h1>
          </br>
          <div class="buttons">
            <button id="${this.prefix}edit-save" class="button">
              <span class="icon is-small"><i class="fas fa-save"></i></span>
            </button>
            <button id="${this.prefix}edit-share" class="button">
              <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
            </button>
            <button id="${this.prefix}edit-close" class="button">
              <span class="icon is-small"><i class="fas fa-times"></i></span>
            </button>
          </div>
        </div>
      </div>
    `;
  }

  async save() {
    try {
      const response = await fetch(this.url, {
        method: "put",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken
        }),
        credentials: "include",
        body: this.gid("edit-content").value
      });
      if (response.status !== 201) {
        throw new Error(`Text content could not be updated (status ${response.status})`);
      }
      content = await response.text();
    } catch (e) {
      Messages.Show("is-warning", e.message);
      console.error(e);
    }
  }
}
