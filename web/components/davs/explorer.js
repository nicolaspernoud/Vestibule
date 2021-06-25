// Imports
import * as Messages from "/services/messages/messages.js";
import * as Auth from "/services/auth/auth.js";
import { AnimateCSS, GetType, Truncate, EncodeURIWithSpecialsCharacters } from "/services/common/common.js";
import { Open } from "/components/davs/open.js";
import { Edit } from "/components/davs/edit.js";
import { Share } from "/components/davs/share.js";
import { Delete } from "/services/common/delete.js";
import { HandleError } from "/services/common/errors.js";

export class Explorer {
  constructor(dav) {
    this.dav = dav;
    this.hostname = dav.host;
    this.fullHostname = `${location.protocol}//${this.hostname}${location.port !== "" ? ":" + location.port : ""}`;
    this.files = [];
    this.path = "/";
    this.encrypted = this.dav.passphrase != null && this.dav.passphrase !== "";
  }

  async mount(mountpoint) {
    const card = document.getElementById(mountpoint);
    card.innerHTML = /* HTML */ `
      <header class="modal-card-head">
        <span class="icon mr-2"> <i class="fas fa-lg fa-${this.dav.icon}" style="color: ${this.dav.color};"></i> </span>
        <p class="modal-card-title">${this.dav.name}</p>
        <button class="delete" aria-label="close" id="explorer-modal-close"></button>
      </header>
      <section id="explorer-modal-content" class="modal-card-body pt-0"></section>
      <progress id="explorer-modal-progress" class="progress is-primary is-small" style="margin-bottom:0px;"></progress>
      <footer id="explorer-modal-footer" class="modal-card-foot">
        <div class="buttons" id="explorer-modal-footer-buttons">
          <button id="explorer-modal-back" class="button is-success">
            <span class="icon is-small">
              <i class="fas fa-arrow-circle-left"></i>
            </span>
          </button>
          ${this.dav.writable
            ? /* HTML */ `
                <button id="explorer-modal-newfolder" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-folder-plus"></i>
                  </span>
                </button>
                <button id="explorer-modal-newtxt" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-file-medical"></i>
                  </span>
                </button>
                <div class="button" style="padding: 0 7px;">
                  <span class="icon is-small" style="margin: 0;">
                    <i class="fas fa-upload"></i>
                  </span>
                  <input class="file-input" type="file" id="explorer-modal-upload" multiple />
                </div>
              `
            : ""}
        </div>
      </footer>
    `;
    this.user = await Auth.GetUser();
    document.getElementById(`explorer-modal-close`).addEventListener("click", async function () {
      const modal = card.parentNode;
      AnimateCSS(modal, "fadeOut");
      await AnimateCSS(card, "zoomOut");
      modal.classList.remove("is-active");
    });
    document.getElementById(`explorer-modal-back`).addEventListener("click", () => {
      this.navigate(goUp(this.path));
    });
    if (this.dav.writable) {
      document.getElementById(`explorer-modal-newfolder`).addEventListener("click", () => {
        this.newFolder();
      });
      document.getElementById(`explorer-modal-newtxt`).addEventListener("click", () => {
        this.newTxt();
      });
      document.getElementById(`explorer-modal-upload`).addEventListener("change", (e) => {
        this.upload(e.target.files);
      });
    }
    this.progress = document.getElementById(`explorer-modal-progress`);
    await this.navigate("/");
  }

  async navigate(destPath) {
    if (!destPath.endsWith("/")) destPath += "/";
    this.path = destPath;
    this.progress.classList.remove("is-hidden");
    try {
      const response = await fetch(this.fullHostname + this.path, {
        method: "PROPFIND",
        headers: new Headers({
          Depth: "1",
          "XSRF-Token": this.user.xsrftoken,
        }),
        credentials: "include",
      });
      if (response.status !== 207) {
        throw new Error(`Files could not be fetched (status ${response.status})`);
      }
      const resXML = await response.text();
      this.files = parseWebDavResponse(resXML);
      this.displayFiles();
      this.progress.classList.add("is-hidden");
    } catch (e) {
      HandleError(e);
    }
  }

  displayFiles() {
    // Create template
    const markup = this.files.map((file) => this.fileTemplate(file)).join("");
    document.getElementById("explorer-modal-content").innerHTML = markup;
    // Register events
    this.files.forEach((file) => {
      this.registerEvents(file);
    });
  }

  fileTemplate(file) {
    return /* HTML */ `
      <article id="file-${file.id}-content" class="media animate__animated animate__fadeIn">
        <figure class="media-left">
          ${file.type.includes("image")
            ? `<p class="image is-48x48"><img id="file-${file.id}-image" src="assets/spinner.svg"/></p>`
            : `<span class="icon is-large"><i class="fas fa-3x fa-${file.isDir ? "folder" : "file"}"></i></span>`}
        </figure>
        <div class="media-content">
          <div class="content">
            <p><strong>${file.name}</strong> <small>(${file.isDir ? "" : sizeToHuman(file.size) + " - "}${intToLocaleDate(file.lastModified)})</small></p>
          </div>
          <nav class="level is-mobile">
            <div class="level-left">
              ${this.encrypted && file.isDir
                ? ""
                : /* HTML */ `
                    <a class="level-item" id="file-${file.id}-download">
                      <span class="icon is-small"><i class="fas fa-download"></i></span>
                    </a>
                  `}
              ${this.dav.writable
                ? /* HTML */ `
                    <a id="file-${file.id}-rename" class="level-item">
                      <span class="icon is-small"><i class="fas fa-pen"></i></span>
                    </a>
                    <a id="file-${file.id}-cut" class="level-item">
                      <span class="icon is-small"><i class="fas fa-cut"></i></span>
                    </a>
                    <a id="file-${file.id}-copy" class="level-item">
                      <span class="icon is-small"><i class="fas fa-copy"></i></span>
                    </a>
                    ${GetType(file) === "text"
                      ? /* HTML */ `
                          <a id="file-${file.id}-edit" class="level-item">
                            <span class="icon is-small"><i class="fas fa-edit"></i></span>
                          </a>
                        `
                      : ""}
                  `
                : ""}
              <a id="file-${file.id}-share" class="level-item">
                <span class="icon is-small"><i class="fas fa-share-alt"></i></span>
              </a>
            </div>
          </nav>
        </div>
        ${this.dav.writable
          ? /* HTML */ `
              <div class="media-right">
                <a id="file-${file.id}-delete">
                  <span class="icon is-small has-text-danger"><i class="fas fa-trash-alt"></i></span>
                </a>
              </div>
            `
          : ""}
      </article>
    `;
  }

  registerEvents(file) {
    if (file.type.includes("image")) {
      LoadImage(document.getElementById(`file-${file.id}-image`), this.fullHostname + file.path, this.user);
    }
    document.getElementById(`file-${file.id}-content`).addEventListener("click", async () => {
      if (file.isDir) {
        this.navigate(file.path);
      } else if (GetType(file) === "document") {
        try {
          const response = await fetch(location.origin + "/api/common/Share", {
            method: "POST",
            headers: new Headers({
              "XSRF-Token": this.user.xsrftoken,
            }),
            credentials: "include",
            body: JSON.stringify({
              sharedfor: "external_editor",
              lifespan: 1,
              // Decode and recode to prevent discrepancies between encoding on browsers
              url: EncodeURIWithSpecialsCharacters(this.hostname + decodeURIComponent(file.path)),
              readonly: false,
            }),
          });
          if (response.status !== 200) {
            throw new Error(`Share token could not be made (status ${response.status})`);
          }
          const token = await response.text();
          window.location.href = `${location.origin}/onlyoffice?file=${this.fullHostname + file.path}&mtime=${new Date(file.lastModified).getTime()}&user=${
            this.user.login
          }&token=${encodeURIComponent(token)}`;
        } catch (e) {
          HandleError(e);
        }
      } else if (GetType(file)) {
        const openModal = new Open(this.user, this.hostname, this.fullHostname, this.files, file);
        openModal.show(true);
      }
    });

    if (this.dav.writable) {
      document.getElementById(`file-${file.id}-rename`).addEventListener("click", (event) => {
        event.stopPropagation();
        this.rename(file);
      });
      document.getElementById(`file-${file.id}-cut`).addEventListener("click", (event) => {
        event.stopPropagation();
        this.moveOrCopy(file, false);
      });
      document.getElementById(`file-${file.id}-copy`).addEventListener("click", (event) => {
        event.stopPropagation();
        this.moveOrCopy(file, true);
      });
      if (GetType(file) === "text") {
        document.getElementById(`file-${file.id}-edit`).addEventListener("click", (event) => {
          event.stopPropagation();
          const editModal = new Edit(this.user, this.fullHostname, file);
          editModal.show(true);
        });
      }
      document.getElementById(`file-${file.id}-delete`).addEventListener("click", (event) => {
        event.stopPropagation();
        this.delete(file);
      });
    }
    if (!(this.encrypted && file.isDir)) {
      document.getElementById(`file-${file.id}-download`).addEventListener("click", (event) => {
        event.stopPropagation();
        this.download(file);
      });
    }
    document.getElementById(`file-${file.id}-share`).addEventListener("click", (event) => {
      event.stopPropagation();
      const shareModal = new Share(this.user, this.hostname, file);
      shareModal.show(true);
    });
  }

  rename(file) {
    let renameModal = document.createElement("div");
    renameModal.classList.add("modal", "animate__animated", "animate__fadeIn", "is-active");
    renameModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box">
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
    const renameOK = renameModal.querySelector("#explorer-rename-ok");
    const renameCancel = renameModal.querySelector("#explorer-rename-cancel");
    const toggleButtons = () => {
      renameOK.classList.toggle("is-loading");
      renameOK.disabled = !renameOK.disabled;
      renameCancel.disabled = !renameCancel.disabled;
    };
    renameOK.addEventListener("click", async () => {
      try {
        toggleButtons();
        const newName = file.isDir ? renameModal.getElementsByTagName("input")[0].value + "/" : renameModal.getElementsByTagName("input")[0].value;
        const response = await fetch(this.fullHostname + file.path, {
          method: "MOVE",
          headers: new Headers({
            Destination: this.fullHostname + path(this.path, newName),
            "XSRF-Token": this.user.xsrftoken,
          }),
          credentials: "include",
        });
        if (response.status !== 201) {
          throw new Error(`File could not be renamed (status ${response.status})`);
        }
        file.name = renameModal.getElementsByTagName("input")[0].value;
        file.path = goUp(file.path) + encodeURIComponent(file.name);
        this.displayFiles();
      } catch (e) {
        HandleError(e);
      }
      await AnimateCSS(renameModal, "fadeOut");
      renameModal.parentNode.removeChild(renameModal);
    });
    renameCancel.addEventListener("click", async () => {
      await AnimateCSS(renameModal, "fadeOut");
      renameModal.parentNode.removeChild(renameModal);
    });
    document.body.appendChild(renameModal);
    field.focus();
  }

  moveOrCopy(file, isCopy) {
    // If there is already a move or copy going, cancel
    if (document.getElementById("explorer-modal-pastecontrol")) {
      return;
    }
    let pasteControl = document.createElement("div");
    pasteControl.id = "explorer-modal-pastecontrol";
    pasteControl.classList.add("field", "has-addons", "animate__animated", "animate__zoomIn", "is-active");
    pasteControl.innerHTML = /* HTML */ `
      <a class="button is-link">
        <span class="icon is-small">
          <i class="fas fa-${isCopy ? "copy" : "cut"}"></i>
        </span>
        <span>Paste <small>"${Truncate(file.name)}"</small> here</span>
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
        const dest = file.isDir ? path(this.path, file.name) + "/" : path(this.path, file.name);
        const response = await fetch(this.fullHostname + file.path, {
          method: isCopy ? "COPY" : "MOVE",
          headers: new Headers({
            Destination: this.fullHostname + dest,
            "XSRF-Token": this.user.xsrftoken,
          }),
          credentials: "include",
        });
        if (response.status !== 201) {
          throw new Error(`File could not be pasted (status ${response.status})`);
        }
        this.navigate(this.path);
      } catch (e) {
        HandleError(e);
      }
      await AnimateCSS(pasteControl, "zoomOut");
      pasteControl.parentNode.removeChild(pasteControl);
    });
    pasteControl.getElementsByTagName("a")[1].addEventListener("click", async () => {
      await AnimateCSS(pasteControl, "zoomOut");
      pasteControl.parentNode.removeChild(pasteControl);
    });
    document.getElementById("explorer-modal-footer-buttons").appendChild(pasteControl);
  }

  async newFolder() {
    const newFolderName = "New Folder";
    const folder = { name: newFolderName, isDir: true, type: "dir", size: 0, lastModified: new Date(), path: path(this.path, newFolderName) + "/", id: this.nextID() };
    try {
      const response = await fetch(this.fullHostname + folder.path, {
        method: "MKCOL",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        credentials: "include",
      });
      if (response.status !== 201) {
        throw new Error(`Folder could not be created (status ${response.status})`);
      }
      this.addFileToView(folder);
    } catch (e) {
      HandleError(e);
    }
  }

  async newTxt() {
    const newTxtName = "New Text.txt";
    const txt = { name: newTxtName, isDir: false, type: "text", size: 0, lastModified: new Date(), path: path(this.path, newTxtName), id: this.nextID() };
    try {
      if (this.files.some((file) => file.name === newTxtName)) {
        throw new Error(`Text document already exists`);
      }
      const response = await fetch(this.fullHostname + txt.path, {
        method: "PUT",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        credentials: "include",
      });
      if (response.status !== 201) {
        throw new Error(`Text document could not be created (status ${response.status})`);
      }
      this.addFileToView(txt);
    } catch (e) {
      HandleError(e);
    }
  }

  delete(file) {
    new Delete(async () => {
      try {
        const response = await fetch(this.fullHostname + file.path, {
          method: "DELETE",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken,
          }),
          credentials: "include",
        });
        if (response.status !== 204) {
          throw new Error(`File could not be deleted (status ${response.status})`);
        }
        this.files = this.files.filter((el) => el.name !== file.name);
        document.getElementById(`file-${file.id}-content`).outerHTML = "";
      } catch (e) {
        HandleError(e);
      }
    }, file.name);
  }

  async upload(files) {
    const onStartPath = this.path;
    let fileIdx = 0;
    for (const file of files) {
      // Check for overwrite
      if (this.files.some((e) => e.name === file.name)) {
        Messages.Show("is-warning", `A file with the name "${file.name}" already exists, please remove the old file before upload.`);
        continue;
      }
      fileIdx++;
      file.path = path(onStartPath, file.name);
      // Create a message to allow progress tracking and cancellation
      let msg = document.createElement("div");
      msg.innerHTML = /* HTML */ `
        <div class="content"><p>${file.name} (file: ${fileIdx}/${files.length})</p></div>
        <progress class="progress is-primary is-small" value="0" max="100" style="margin-bottom: 0px;"></progress>
      `;
      msg.classList.add("is-info", "notification", "uploader", "animate__animated", "animate__fadeInUp");
      const delBtn = document.createElement("button");
      let xhr = new XMLHttpRequest();
      // track upload progress
      xhr.upload.onprogress = function (e) {
        msg.getElementsByTagName("progress")[0].value = (e.loaded / e.total) * 100;
      };
      delBtn.addEventListener("click", async () => {
        xhr.abort();
        try {
          const response = await fetch(this.fullHostname + file.path, {
            method: "DELETE",
            headers: new Headers({
              "XSRF-Token": this.user.xsrftoken,
            }),
            credentials: "include",
          });
          if (response.status !== 204) {
            throw new Error(`Cancelled file could not be deleted (status ${response.status})`);
          }
        } catch (e) {
          console.error(e);
        }
      });
      delBtn.classList.add("delete");
      msg.appendChild(delBtn);
      document.body.appendChild(msg);
      try {
        await this.uploadFile(xhr, file);
        if (this.path === onStartPath) {
          const newFile = { name: file.name, path: file.path, isDir: file.isDir, type: file.type, size: file.size, lastModified: file.lastModified, id: this.nextID() };
          this.addFileToView(newFile);
        }
      } catch (e) {
        console.error(e.statusText);
        Messages.Show("is-warning", e.statusText);
      }
      await AnimateCSS(msg, "fadeOutDown");
      msg.parentNode.removeChild(msg);
    }
  }

  addFileToView(file) {
    this.files.push(file);
    const markup = this.fileTemplate(file);
    try {
      document.getElementById(`file-${file.id - 1}-content`).insertAdjacentHTML("afterend", markup);
    } catch {
      document.getElementById(`explorer-modal-content`).innerHTML = markup;
    }
    this.registerEvents(file);
  }

  uploadFile(xhr, file) {
    return new Promise((resolve, reject) => {
      xhr.withCredentials = true;
      // track completion: both successful or not
      xhr.onloadend = () => {
        if (xhr.status === 0) {
          reject({
            status: xhr.status,
            statusText: `Upload of ${file.name} cancelled`,
          });
        } else if (xhr.status == 201) {
          resolve(xhr.status);
        } else {
          reject({
            status: xhr.status,
            statusText: `Error uploading ${file.name} (status ${xhr.status})`,
          });
        }
      };
      xhr.onerror = function (e) {
        reject({
          status: this.status,
          statusText: `Error uploading ${file.name} (status ${xhr.status})`,
        });
      };
      xhr.open("PUT", this.fullHostname + file.path);
      xhr.setRequestHeader("XSRF-Token", this.user.xsrftoken);
      xhr.setRequestHeader("X-OC-Mtime", Math.round(file.lastModified / 1000));
      xhr.send(file);
    });
  }

  async download(file) {
    try {
      const response = await fetch(location.origin + "/api/common/Share", {
        method: "POST",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        credentials: "include",
        body: JSON.stringify({
          sharedfor: "download",
          lifespan: 1,
          url: this.hostname + file.path,
          readonly: true,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Share token could not be made (status ${response.status})`);
      }
      const shareToken = await response.text();
      const shareURL = `${this.fullHostname + file.path}?token=${encodeURIComponent(shareToken)}`;
      const link = document.createElement("a");
      link.href = shareURL;
      link.click();
    } catch (e) {
      HandleError(e);
    }
  }

  nextID() {
    if (!this.files.length) {
      return 1;
    }
    return this.files.reduce((a, b) => (a.id > b.id ? a : b)).id + 1;
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
    file.id = i;
    files.push(file);
  }
  return files.sort(fileSortFunction);
}

function goUp(destPath) {
  if (destPath === "/") return destPath;
  if (destPath.endsWith("/")) destPath = destPath.substring(0, destPath.length - 1);
  const lastSlashPosition = destPath.lastIndexOf("/");
  return lastSlashPosition === 0 ? "/" : destPath.substring(0, lastSlashPosition + 1);
}

function sizeToHuman(size) {
  const i = size === 0 ? 0 : Math.floor(Math.log(size) / Math.log(1024));
  return (size / Math.pow(1024, i)).toFixed(2) + " " + ["B", "kB", "MB", "GB", "TB"][i];
}

function intToLocaleDate(idate) {
  return new Date(idate).toLocaleString();
}

function fileSortFunction(a, b) {
  if (a.isDir !== b.isDir) {
    if (a.isDir) {
      return -1;
    } else {
      return 1;
    }
  } else {
    return a.name.localeCompare(b.name);
  }
}

export async function LoadImage(image, url, user) {
  try {
    const response = await fetch(url, {
      method: "GET",
      headers: new Headers({
        "XSRF-Token": user.xsrftoken,
      }),
      credentials: "include",
    });
    if (response.status !== 200) {
      throw new Error(`Error loading image (status ${response.status})`);
    }
    const blob = await response.blob();
    const objectURL = URL.createObjectURL(blob);
    image.src = objectURL;
  } catch (e) {
    HandleError(e);
  }
}

function path(destPath, name) {
  return destPath + encodeURIComponent(name);
}
