// Imports
import { Icons } from "/services/common/icons.js";
import { AnimateCSS, RandomString, IsEmpty } from "/services/common/common.js";
import { Explorer } from "./explorer.js";
import { Delete } from "/services/common/delete.js";
import { GetColor } from "../sysinfo/sysinfo.js";
import { HandleError } from "/services/common/errors.js";

export async function mount(where, user) {
  const davsComponent = new Davs(user);
  await davsComponent.mount(where);
}

class Davs {
  constructor(user) {
    this.user = user;
  }

  // DOM elements
  id_field;
  name_field;
  icon_field;
  color_field;
  host_field;
  writable_field;
  root_field;
  secured_field;
  roles_field;
  roles_container;
  passphrase_field;

  async mount(mountpoint) {
    document.getElementById(mountpoint).innerHTML = /* HTML */ `
      <div id="davs-list" class="flex-container"></div>
      <button id="davs-new" class="button is-primary is-hidden">
        <span class="icon is-small">
          <i class="fas fa-plus"></i>
        </span>
      </button>

      <div class="modal" id="davs-modal">
        <div class="modal-background"></div>
        <div class="modal-card" id="davs-modal-card">
          <header class="modal-card-head">
            <p class="modal-card-title">Add/Edit dav</p>
            <button class="delete" aria-label="close" id="davs-modal-close"></button>
          </header>
          <section class="modal-card-body">
            <div class="field">
              <label class="label">Id (changing the id will overwrite a dav with the same id, editing a dav and changing for an unused id will clone the dav)</label>
              <div class="control">
                <input class="input" type="number" id="davs-modal-id" />
              </div>
            </div>
            <div class="field">
              <label class="label">Name</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-name" />
              </div>
            </div>
            <div class="field">
              <label class="label">Icon</label>
              <div class="control has-icons-left">
                <input class="input" type="text" id="davs-modal-icon" />
                <span class="icon is-small is-left has-text-info">
                  <i id="davs-modal-icon-preview" class="fas fa-file"></i>
                </span>
              </div>
            </div>
            <div class="field">
              <label class="label">Color</label>
              <div class="control">
                <input class="input" type="color" id="davs-modal-color" />
              </div>
            </div>
            <div class="field">
              <label class="label">Host</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-host" />
              </div>
            </div>
            <div class="field">
              <div class="control">
                <label class="label"><input id="davs-modal-writable" type="checkbox" />Allow write access</label>
              </div>
            </div>
            <div class="field" id="davs-modal-root-container">
              <label class="label">Root directory to serve</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-root" />
              </div>
            </div>
            <div class="field">
              <div class="control">
                <label class="label"><input id="davs-modal-secured" type="checkbox" />Secure access</label>
              </div>
            </div>
            <div class="field" id="davs-modal-roles-container">
              <label class="label">Allow access to roles (separated with commas)</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-roles" />
              </div>
            </div>
            <label class="label">Passphrase to encrypt files (leave empty to disable encryption)</label>
            <div class="field has-addons">
              <div class="control is-expanded">
                <input class="input" type="text" id="davs-modal-passphrase" />
              </div>
              <div class="control">
                <button id="davs-modal-passphrase-remove" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-times"></i>
                  </span>
                </button>
              </div>
              <div class="control">
                <button id="davs-modal-passphrase-generate" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-dice"></i>
                  </span>
                </button>
              </div>
            </div>
            <br />
          </section>
          <footer class="modal-card-foot">
            <button id="davs-modal-save" class="button is-success">Save changes</button>
            <button id="davs-modal-cancel" class="button">Cancel</button>
          </footer>
        </div>
      </div>

      <div class="modal animate__animated animate__zoomIn" id="davs-icons-modal">
        <div class="modal-card">
          <section id="davs-icons-modal-list" class="modal-card-body"></section>
        </div>
      </div>

      <div class="modal" id="davs-explorer-modal">
        <div class="modal-background"></div>
        <div id="davs-explorer-modal-card" class="modal-card"></div>
      </div>
    `;
    if (!IsEmpty(this.user)) {
      if (this.user.isAdmin) document.getElementById("davs-new").classList.toggle("is-hidden");
      this.registerModalFields();
      await this.firstShowDavs();
    }
  }

  davTemplate(dav) {
    this.cleanDav(dav);
    const du = dav.usedgb / dav.totalgb;
    const free = dav.totalgb - dav.usedgb;
    return /* HTML */ `
      <div id="davs-dav-${dav.id}" class="card icon-card">
        <div id="davs-dav-open-${dav.id}" class="card-content has-text-centered">
          <span class="icon is-medium" style="color:${dav.color};">
            <i class="fas fa-3x fa-${dav.icon ? dav.icon : "file"}"></i>
          </span>
        </div>
        <p class="has-text-centered"><strong>${dav.name ? dav.name : dav.id}</strong></p>
        <div class="card-footer">
          <div class="dropdown is-hoverable" style="margin-top: 1px;">
            <div class="dropdown-trigger">
              <button class="button is-white">
                <span class="icon is-small">
                  <i class="fas fa-angle-down"></i>
                </span>
              </button>
            </div>
            <div class="dropdown-menu animate__animated animate__fadeIn" role="menu">
              <div class="dropdown-content">
                <div class="dropdown-item"></div>
                ${this.user.isAdmin ? '<a class="dropdown-item" id="davs-dav-edit-' + dav.id + '"><i class="fas fa-edit"></i><strong> Edit</strong></a>' : ""}
                ${this.user.isAdmin
                  ? '<a class="dropdown-item has-text-danger" id="davs-dav-delete-' + dav.id + '"><i class="fas fa-trash-alt"></i><strong> Delete</strong></a>'
                  : ""}
                <hr class="dropdown-divider" />
                <div class="dropdown-item">
                  <p>
                    <progress class="progress is-${GetColor(du)} is-small" value="${dav.usedgb}" max="${dav.totalgb}"></progress>${dav.usedgb !== undefined
                      ? free + " GB free"
                      : ""}
                  </p>
                  <hr class="dropdown-divider" />
                  <p><strong>${dav.host}</strong></p>
                  <p>Serves ${dav.root} directory, with ${dav.writable ? "read/write" : "read only"} access</p>
                  <p>${dav.secured ? "Restricted access to users with roles <strong>" + dav.roles + "</strong>" : "Unrestricted access"}</p>
                  <p class="has-text-centered"><strong>-${dav.id}-</strong></p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  displayDavs(inDavs) {
    const markup = inDavs
      .map((dav) => {
        if (this.user.isAdmin || !dav.secured || dav.roles.some((r) => this.user.memberOf.includes(r))) {
          return this.davTemplate(dav);
        }
      })
      .join("");
    document.getElementById("davs-list").innerHTML = markup;
    inDavs.map((dav) => {
      if (this.user.isAdmin) {
        document.getElementById(`davs-dav-edit-${dav.id}`).addEventListener("click", () => {
          this.editDav(dav);
        });
        document.getElementById(`davs-dav-delete-${dav.id}`).addEventListener("click", () => {
          new Delete(() => {
            this.deleteDav(dav);
          }, dav.name);
        });
      }
      if (this.user.isAdmin || !dav.secured || dav.roles.some((r) => this.user.memberOf.includes(r))) {
        document.getElementById(`davs-dav-open-${dav.id}`).addEventListener("click", () => {
          this.openExplorerModal(dav);
        });
      }
    });
  }

  async firstShowDavs() {
    try {
      const response = await fetch("/api/common/davs", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Davs could not be fetched (status ${response.status})`);
      }
      this.davs = await response.json();
      this.displayDavs(this.davs);
    } catch (e) {
      HandleError(e);
    }
  }

  async deleteDav(dav) {
    try {
      const response = await fetch("/api/admin/davs/" + dav.id, {
        method: "delete",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Dav could not be deleted (status ${response.status})`);
      }
      this.reloadDavsOnServer();
      document.getElementById(`davs-dav-${dav.id}`).remove();
    } catch (e) {
      HandleError(e);
    }
  }

  registerModalFields() {
    this.id_field = document.getElementById("davs-modal-id");
    this.name_field = document.getElementById("davs-modal-name");
    this.icon_field = document.getElementById("davs-modal-icon");
    this.color_field = document.getElementById("davs-modal-color");
    this.host_field = document.getElementById("davs-modal-host");
    this.writable_field = document.getElementById("davs-modal-writable");
    this.root_field = document.getElementById("davs-modal-root");
    this.secured_field = document.getElementById("davs-modal-secured");
    this.roles_field = document.getElementById("davs-modal-roles");
    this.roles_container = document.getElementById("davs-modal-roles-container");
    this.passphrase_field = document.getElementById("davs-modal-passphrase");
    document.getElementById(`davs-modal-close`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`davs-modal-cancel`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`davs-modal-save`).addEventListener("click", () => {
      this.postDav();
    });
    document.getElementById(`davs-new`).addEventListener("click", () => {
      this.newDav();
    });
    this.icon_field.addEventListener("click", () => {
      this.pickIcon();
    });
    this.secured_field.addEventListener("click", () => {
      this.toggleRoles();
    });
    document.getElementById(`davs-modal-passphrase-generate`).addEventListener("click", () => {
      this.passphrase_field.value = RandomString(48);
    });
    document.getElementById(`davs-modal-passphrase-remove`).addEventListener("click", () => {
      this.passphrase_field.value = "";
    });
  }

  async editDav(dav) {
    this.cleanDav(dav);
    this.id_field.value = dav.id;
    this.name_field.value = dav.name;
    this.icon_field.value = dav.icon;
    this.color_field.value = dav.color;
    this.host_field.value = dav.host;
    this.writable_field.checked = dav.writable;
    this.root_field.value = dav.root;
    this.secured_field.checked = dav.secured;
    this.roles_field.value = dav.roles;
    this.passphrase_field.value = dav.passphrase;
    this.toggleModal();
  }

  cleanDav(dav) {
    let props = ["writable", "name", "roles", "passphrase"];
    for (const prop of props) {
      dav[prop] = dav[prop] === undefined ? "" : dav[prop];
    }
    dav.icon = dav.icon === undefined ? "file" : dav.icon;
  }

  async newDav() {
    let maxid = 0;
    this.davs.map(function (dav) {
      if (dav.id > maxid) maxid = dav.id;
    });
    this.id_field.value = maxid + 1;
    this.name_field.value = "";
    this.icon_field.value = "file";
    this.color_field.value = "#000000";
    this.host_field.value = `new_dav_service.${location.hostname}`;
    this.writable_field.checked = false;
    this.root_field.value = "";
    this.secured_field.checked = false;
    this.roles_field.value = "";
    this.passphrase_field.value = "";
    this.toggleModal();
  }

  async postDav() {
    try {
      const response = await fetch("/api/admin/davs/", {
        method: "post",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        body: JSON.stringify({
          id: parseInt(this.id_field.value),
          name: this.name_field.value,
          icon: this.icon_field.value,
          color: this.color_field.value,
          host: this.host_field.value,
          writable: this.writable_field.checked,
          root: this.root_field.value,
          secured: this.secured_field.checked,
          roles: this.secured_field.checked ? this.roles_field.value.split(",") : "",
          passphrase: this.passphrase_field.value,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Davs could not be updated (status ${response.status})`);
      }
      this.davs = await response.json();
      this.displayDavs(this.davs);
      await this.reloadDavsOnServer();
    } catch (e) {
      HandleError(e);
    }
    this.toggleModal();
  }

  async reloadDavsOnServer() {
    try {
      const response = await fetch("/api/admin/reload", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Dav could not be reloaded (status ${response.status})`);
      }
    } catch (e) {
      HandleError(e);
    }
  }

  async toggleModal() {
    this.toggleRoles();
    this.updateIcon();
    const modal = document.getElementById("davs-modal");
    const card = document.getElementById("davs-modal-card");
    if (modal.classList.contains("is-active")) {
      AnimateCSS(modal, "fadeOut");
      await AnimateCSS(card, "zoomOut");
      modal.classList.remove("is-active");
    } else {
      modal.classList.add("is-active");
      AnimateCSS(modal, "fadeIn");
      AnimateCSS(card, "zoomIn");
    }
  }

  toggleRoles() {
    if (this.secured_field.checked) {
      this.roles_container.style.display = "block";
    } else {
      this.roles_container.style.display = "none";
    }
  }

  updateIcon() {
    document.getElementById("davs-modal-icon-preview").setAttribute("class", "fas fa-" + this.icon_field.value);
  }

  async pickIcon() {
    const iconsTemplate =
      '<div class="buttons">' +
      Icons.map(
        (icon) => /* HTML */ `
          <button class="button${this.icon_field.value == icon ? " is-primary" : ""}" id="davs-icon-modal-list-${icon}">
            <span class="icon">
              <i class="fas fa-${icon}"></i>
            </span>
          </button>
        `
      ).join("") +
      "</div>";
    document.getElementById("davs-icons-modal-list").innerHTML = iconsTemplate;
    Icons.map((icon) =>
      document.getElementById(`davs-icon-modal-list-${icon}`).addEventListener("click", () => {
        this.icon_field.value = icon;
        this.updateIcon();
        document.getElementById("davs-icons-modal").classList.toggle("is-active");
      })
    );
    document.getElementById("davs-icons-modal").classList.toggle("is-active");
  }

  openExplorerModal(dav) {
    const modal = document.getElementById("davs-explorer-modal");
    const card = document.getElementById("davs-explorer-modal-card");
    const explorer = new Explorer(dav);
    explorer.mount("davs-explorer-modal-card");
    modal.classList.add("is-active");
    AnimateCSS(modal, "fadeIn");
    AnimateCSS(card, "zoomIn");
  }
}
