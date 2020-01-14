// Imports
import * as Messages from "/services/messages/messages.js";
import * as Auth from "/services/auth/auth.js";
import { AnimateCSS } from "/services/common/common.js";

export class Explorer {
  constructor(hostname) {
    this.hostname = `${location.protocol}//${hostname}${location.port !== "" ? ":" + location.port : ""}`;
    this.files = [];
    this.path = "/";
  }

  async mount(mountpoint, readwrite) {
    this.readwrite = readwrite;
    const card = document.getElementById(mountpoint);
    card.innerHTML = /* HTML */ `
      <header class="modal-card-head">
        <p class="modal-card-title">Explorer</p>
        <button class="delete" aria-label="close" id="explorer-modal-close"></button>
      </header>
      <section id="explorer-modal-content" class="modal-card-body"></section>
      <progress id="explorer-modal-progress" class="progress is-small is-primary" style="margin-bottom:0px;"></progress>
      <footer id="explorer-modal-footer" class="modal-card-foot">
        <div class="buttons" id="explorer-modal-footer-buttons">
          <button id="explorer-modal-back" class="button is-success">
            <span class="icon is-small">
              <i class="fas fa-arrow-circle-left"></i>
            </span>
          </button>
          ${this.readwrite
            ? /* HTML */ `
                <button id="explorer-modal-newfolder" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-folder-plus"></i>
                  </span>
                </button>
                <button class="button" style="padding-left: 5px; padding-right: 5px;">
                  <input id="explorer-modal-upload" class="file-input" type="file" multiple />
                  <span class="icon is-small is-marginless">
                    <i class="fas fa-upload"></i>
                  </span>
                </button>
              `
            : ""}
        </div>
      </footer>
    `;
    this.user = await Auth.GetUser();
    document.getElementById(`explorer-modal-close`).addEventListener("click", function() {
      const modal = card.parentNode;
      AnimateCSS(modal, "fadeOut");
      AnimateCSS(card, "zoomOut", function() {
        modal.classList.remove("is-active");
      });
    });
    document.getElementById(`explorer-modal-back`).addEventListener("click", () => {
      this.navigate(goUp(this.path));
    });
    if (this.readwrite) {
      document.getElementById(`explorer-modal-newfolder`).addEventListener("click", () => {
        this.newFolder();
      });
      document.getElementById(`explorer-modal-upload`).addEventListener("change", e => {
        this.upload(e.srcElement.files);
      });
    }
    this.progress = document.getElementById(`explorer-modal-progress`);
    await this.navigate("/");
  }

  async navigate(path) {
    this.path = path;
    this.progress.classList.remove("is-hidden");
    try {
      const response = await fetch(this.hostname + this.path, {
        method: "PROPFIND",
        headers: new Headers({
          Depth: "1",
          "XSRF-Token": this.user.xsrftoken
        }),
        credentials: "include"
      });
      if (response.status !== 207) {
        throw new Error(`Files could not be fetched (status ${response.status})`);
      }
      const resXML = await response.text();
      this.files = parseWebDavResponse(resXML);
      this.displayFiles();
      this.progress.classList.add("is-hidden");
    } catch (e) {
      Messages.Show("is-warning", e.message);
      console.error(e);
    }
  }

  displayFiles() {
    const content = document.getElementById("explorer-modal-content");
    while (content.firstChild) {
      content.removeChild(content.firstChild);
    }
    this.files.map(file => {
      content.appendChild(this.fileTemplate(file));
    });
  }

  fileTemplate(file) {
    let el = document.createElement("article");
    el.classList.add("media", "animated", "fadeIn", "faster");
    el.innerHTML = /* HTML */ `
      <figure class="media-left">
        ${file.type.includes("image")
          ? '<p class="image is-48x48"><img src="' + this.hostname + file.path + '"/></p>'
          : `<span class="icon is-large"><i class="fas fa-3x fa-${file.isDir ? "folder" : "file"}"></i></span>`}
      </figure>
      <div class="media-content">
        <div id="explorer-content" class="content">
          <p><strong>${file.name}</strong> <small>(${file.isDir ? "" : sizeToHuman(file.size) + " - "}${intToLocaleDate(file.lastModified)})</small></p>
        </div>
        <nav class="level is-mobile">
          <div class="level-left">
            <a class="level-item" href=${this.hostname + file.path}>
              <span class="icon is-small"><i class="fas fa-download"></i></span>
            </a>
            ${this.readwrite
              ? /* HTML */ `
                  <a id="explorer-rename" class="level-item">
                    <span class="icon is-small"><i class="fas fa-pen"></i></span>
                  </a>
                  <a id="explorer-cut" class="level-item">
                    <span class="icon is-small"><i class="fas fa-cut"></i></span>
                  </a>
                  <a id="explorer-copy" class="level-item">
                    <span class="icon is-small"><i class="fas fa-copy"></i></span>
                  </a>
                `
              : ""}
            <a id="explorer-share" class="level-item">
              <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
            </a>
          </div>
        </nav>
      </div>
      ${this.readwrite
        ? /* HTML */ `
            <div class="media-right">
              <a id="explorer-delete">
                <span class="icon is-small has-text-danger"><i class="fas fa-trash-alt"></i></span>
              </a>
            </div>
          `
        : ""}
    `;
    if (file.isDir) {
      el.querySelector("#" + "explorer-content").addEventListener("click", () => {
        this.navigate(file.path);
      });
    }
    if (this.readwrite) {
      el.querySelector("#" + "explorer-rename").addEventListener("click", () => {
        this.rename(file);
      });
      el.querySelector("#" + "explorer-cut").addEventListener("click", () => {
        this.moveOrCopy(file, false);
      });
      el.querySelector("#" + "explorer-copy").addEventListener("click", () => {
        this.moveOrCopy(file, true);
      });
      el.querySelector("#" + "explorer-delete").addEventListener("click", () => {
        this.delete(file);
      });
    }
    el.querySelector("#" + "explorer-share").addEventListener("click", () => {
      this.share(file);
    });
    return el;
  }

  rename(file) {
    let renameModal = document.createElement("div");
    renameModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
    renameModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box" style="margin: 2rem;">
          <div class="field">
            <label class="label">Name</label>
            <div class="control">
              <input class="input" type="text" value="${file.name}" />
            </div>
          </div>
          <div class="field is-grouped">
            <div class="control">
              <button id="explorer-rename-ok" class="button is-success">
                <span class="icon is-small"><i class="fas fa-check"></i></span>
              </button>
            </div>
            <div class="control">
              <button id="explorer-rename-cancel" class="button is-danger">
                <span class="icon is-small"><i class="fas fa-times-circle"></i></span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
    const field = renameModal.getElementsByTagName("input")[0];
    field.setSelectionRange(0, file.name.lastIndexOf("."));
    renameModal.querySelector("#" + "explorer-rename-ok").addEventListener("click", async () => {
      try {
        const response = await fetch(this.hostname + file.path, {
          method: "MOVE",
          headers: new Headers({
            Destination: this.hostname + this.path + renameModal.getElementsByTagName("input")[0].value,
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include"
        });
        if (response.status !== 201) {
          throw new Error(`File could not be renamed (status ${response.status})`);
        }
        file.name = renameModal.getElementsByTagName("input")[0].value;
        file.path = goUp(file.path) + file.name;
        this.displayFiles();
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
      AnimateCSS(renameModal, "fadeOut", function() {
        renameModal.parentNode.removeChild(renameModal);
      });
    });
    renameModal.querySelector("#" + "explorer-rename-cancel").addEventListener("click", () => {
      AnimateCSS(renameModal, "fadeOut", function() {
        renameModal.parentNode.removeChild(renameModal);
      });
    });
    document.body.appendChild(renameModal);
    field.focus();
  }

  moveOrCopy(file, isCopy) {
    let pasteControl = document.createElement("div");
    pasteControl.classList.add("field", "has-addons", "animated", "zoomIn", "faster", "is-active");
    pasteControl.innerHTML = /* HTML */ `
      <a class="button is-link">
        <span class="icon is-small">
          <i class="fas fa-${isCopy ? "copy" : "cut"}"></i>
        </span>
        <span>Paste ${file.name} here</span>
      </a>
      <a class="button is-link">
        <span class="icon is-small">
          <i class="fas fa-times"></i>
        </span>
        <span>Cancel</span>
      </a>
    `;
    pasteControl.getElementsByTagName("a")[0].addEventListener("click", async () => {
      try {
        const response = await fetch(this.hostname + file.path, {
          method: isCopy ? "COPY" : "MOVE",
          headers: new Headers({
            Destination: this.hostname + this.path + file.name,
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include"
        });
        if (response.status !== 201) {
          throw new Error(`File could not be pasted (status ${response.status})`);
        }
        this.navigate(this.path);
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
      AnimateCSS(pasteControl, "zoomOut", function() {
        pasteControl.parentNode.removeChild(pasteControl);
      });
    });
    pasteControl.getElementsByTagName("a")[1].addEventListener("click", async () => {
      AnimateCSS(pasteControl, "zoomOut", function() {
        pasteControl.parentNode.removeChild(pasteControl);
      });
    });
    document.getElementById("explorer-modal-footer-buttons").appendChild(pasteControl);
  }

  async newFolder() {
    const folder = { name: "New Folder", isDir: true, type: "dir", size: 0, lastModified: new Date(), path: this.path + "New Folder" };
    try {
      const response = await fetch(this.hostname + folder.path, {
        method: "MKCOL",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken
        }),
        credentials: "include"
      });
      if (response.status !== 201) {
        throw new Error(`Folder could not be created (status ${response.status})`);
      }
      this.files.push(folder);
      this.displayFiles();
    } catch (e) {
      Messages.Show("is-warning", e.message);
      console.error(e);
    }
  }

  delete(file) {
    let deleteModal = document.createElement("div");
    deleteModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
    deleteModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box" style="margin: 2rem;">
          <div class="field">
            <label class="label">Confirm</label>
          </div>
          <div class="field is-grouped">
            <div class="control">
              <button id="explorer-delete-ok" class="button is-danger">
                <span class="icon"><i class="fas fa-check"></i></span><span>Delete</span>
              </button>
            </div>
            <div class="control">
              <button id="explorer-delete-cancel" class="button">
                <span class="icon"><i class="fas fa-times-circle"></i></span><span>Cancel</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
    deleteModal.querySelector("#" + "explorer-delete-ok").addEventListener("click", async () => {
      try {
        const response = await fetch(this.hostname + file.path, {
          method: "DELETE",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include"
        });
        if (response.status !== 204) {
          throw new Error(`File could not be deleted (status ${response.status})`);
        }
        this.files = this.files.filter(el => el.name !== file.name);
        this.displayFiles();
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
      AnimateCSS(deleteModal, "fadeOut", function() {
        deleteModal.parentNode.removeChild(deleteModal);
      });
    });
    deleteModal.querySelector("#" + "explorer-delete-cancel").addEventListener("click", () => {
      AnimateCSS(deleteModal, "fadeOut", function() {
        deleteModal.parentNode.removeChild(deleteModal);
      });
    });
    document.body.appendChild(deleteModal);
  }

  upload(files) {
    const onStartPath = this.path;
    let offset = 0;
    for (const file of files) {
      file.path = this.path + file.name;
      // Create a message to allow progress tracking and cancellation
      let msg = document.createElement("div");
      msg.innerHTML = /* HTML */ `
        <div class="content"><p>${file.name}</p></div>
        <progress class="progress is-primary" value="0" max="100" style="margin-bottom: 0px;"></progress>
      `;
      msg.classList.add("is-info", "notification", "uploader", "animated", "fadeInUp", "faster");
      msg.style.marginBottom = offset.toString() + "px";
      const delBtn = document.createElement("button");
      delBtn.classList.add("delete");
      msg.appendChild(delBtn);
      // Perform the request
      let xhr = new XMLHttpRequest();
      xhr.withCredentials = true;
      // track upload progress
      xhr.upload.onprogress = function(e) {
        msg.getElementsByTagName("progress")[0].value = (e.loaded / e.total) * 100;
      };
      // track completion: both successful or not
      xhr.onloadend = () => {
        if (xhr.status === 0) {
          console.log(`Upload of ${file.name} cancelled`);
        } else if (xhr.status == 201) {
          if (this.path === onStartPath) {
            this.files.push(file);
            this.displayFiles();
          }
        } else {
          const message = `Error uploading ${file.name} (status ${xhr.status})`;
          Messages.Show("is-warning", message);
          console.error(message);
        }
        AnimateCSS(msg, "fadeOutDown", function() {
          msg.parentNode.removeChild(msg);
        });
      };
      xhr.onerror = function(e) {
        Messages.Show("is-warning", e.message);
      };
      xhr.open("PUT", this.hostname + file.path);
      xhr.send(file);
      delBtn.addEventListener("click", async () => {
        xhr.abort();
        try {
          const response = await fetch(this.hostname + file.path, {
            method: "DELETE",
            headers: new Headers({
              "XSRF-Token": this.user.xsrftoken
            }),
            credentials: "include"
          });
          if (response.status !== 204) {
            throw new Error(`Cancelled file could not be deleted (status ${response.status})`);
          }
        } catch (e) {
          console.error(e);
        }
      });
      document.body.appendChild(msg);
      offset = offset + 50;
    }
  }

  share(file) {
    let shareModal = document.createElement("div");
    shareModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
    shareModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box" style="margin: 2rem;">
          <div class="field">
            <label class="label">Share with</label>
            <div class="control">
              <input id="explorer-share-for" class="input" type="text" />
            </div>
          </div>
          <div class="field">
            <label class="label">Days</label>
            <div class="control">
              <input id="explorer-share-howlong" class="input" type="number" value="7" />
            </div>
          </div>
          <div class="field is-grouped">
            <div class="control">
              <button id="explorer-share-ok" class="button is-success">
                <span class="icon is-small"><i class="fas fa-check"></i></span>
              </button>
            </div>
            <div class="control">
              <button id="explorer-share-cancel" class="button is-danger">
                <span class="icon is-small"><i class="fas fa-times-circle"></i></span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
    shareModal.querySelector("#" + "explorer-share-ok").addEventListener("click", async () => {
      try {
        const lifespan = parseInt(shareModal.querySelector("#" + "explorer-share-howlong").value);
        const response = await fetch(location.origin + "/api/common/Share", {
          method: "POST",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include",
          body: JSON.stringify({
            sharedfor: shareModal.querySelector("#" + "explorer-share-for").value,
            lifespan: lifespan,
            url: this.hostname + file.path,
            readonly: true
          })
        });
        if (response.status !== 200) {
          throw new Error(`Share token could not be made (status ${response.status})`);
        }
        const shareToken = await response.text();
        // Create result modal
        let resultModal = document.createElement("div");
        resultModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
        resultModal.innerHTML = /* HTML */ `
          <div class="modal-background"></div>
          <div class="modal-content">
            <div class="box" style="margin: 2rem;">
              <div class="content is-small">
                <h1>This link will be available during ${lifespan} days</h1>
                <a href="${this.hostname + file.path + "?token=" + shareToken}" class="button">
                  <span class="icon">
                    <i class="fas fa-link"></i>
                  </span>
                  <span>Download</span>
                </a>
              </div>
              <div class="field is-grouped">
                <div class="control">
                  <button id="explorer-result-close" class="button is-success">
                    <span class="icon is-small"><i class="fas fa-check"></i></span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        `;
        resultModal.querySelector("#" + "explorer-result-close").addEventListener("click", () => {
          AnimateCSS(resultModal, "fadeOut", function() {
            resultModal.parentNode.removeChild(resultModal);
          });
        });
        document.body.appendChild(resultModal);
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
      AnimateCSS(shareModal, "fadeOut", function() {
        shareModal.parentNode.removeChild(shareModal);
      });
    });
    shareModal.querySelector("#" + "explorer-share-cancel").addEventListener("click", () => {
      AnimateCSS(shareModal, "fadeOut", function() {
        shareModal.parentNode.removeChild(shareModal);
      });
    });
    document.body.appendChild(shareModal);
  }
}

function parseWebDavResponse(txt) {
  const parser = new DOMParser();
  const xmlDoc = parser.parseFromString(txt, "text/xml");
  const files = [];

  const x = xmlDoc.getElementsByTagName("D:response");
  // Start from 1 to remove root
  for (let i = 1; i < x.length; i++) {
    const file = {};
    file.name = x[i].getElementsByTagName("D:displayname")[0].textContent;
    file.path = x[i].getElementsByTagName("D:href")[0].textContent;
    file.isDir = x[i].getElementsByTagName("D:resourcetype")[0].hasChildNodes();
    file.type = file.isDir ? "dir" : x[i].getElementsByTagName("D:getcontenttype")[0].textContent;
    file.size = file.isDir ? 0 : parseInt(x[i].getElementsByTagName("D:getcontentlength")[0].textContent);
    file.lastModified = x[i].getElementsByTagName("D:getlastmodified")[0].textContent;
    files.push(file);
  }
  return files;
}

function goUp(path) {
  if (path === "/") return path;
  if (path.endsWith("/")) path = path.substring(0, path.length - 1);
  const lastSlashPosition = path.lastIndexOf("/");
  return lastSlashPosition === 0 ? "/" : path.substring(0, lastSlashPosition);
}

function sizeToHuman(size) {
  const i = size === 0 ? 0 : Math.floor(Math.log(size) / Math.log(1024));
  return (size / Math.pow(1024, i)).toFixed(2) + " " + ["B", "kB", "MB", "GB", "TB"][i];
}

function intToLocaleDate(idate) {
  return new Date(idate).toLocaleString();
}
