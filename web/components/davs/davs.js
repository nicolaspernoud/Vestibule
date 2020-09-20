// Imports
import * as Auth from "/services/auth/auth.js";
import { Icons } from "/services/common/icons.js";
import { AnimateCSS, RandomString } from "/services/common/common.js";
import { Explorer } from "./explorer.js";
import { Delete } from "/services/common/delete.js";
import { GetColor } from "../sysinfo/sysinfo.js";
import { HandleError } from "/services/common/errors.js";

// DOM elements
let mountpoint;
let id_field;
let name_field;
let icon_field;
let color_field;
let host_field;
let writable_field;
let root_field;
let secured_field;
let roles_field;
let roles_container;
let passphrase_field;
let root_container;
let iss3_field;
let s3_container;
let endpoint_field;
let region_field;
let bucket_field;
let accesskeyid_field;
let secretaccesskey_field;

// local variables
let davs;
let user;

export async function mount(where) {
  mountpoint = where;
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
            <label class="label">Id</label>
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

          <div class="field">
            <div class="control">
              <label class="label"><input id="davs-modal-iss3" type="checkbox" />Use S3 compatible backend</label>
            </div>
          </div>

          <div class="field" id="davs-modal-root-container">
            <label class="label">Root directory to serve</label>
            <div class="control">
              <input class="input" type="text" id="davs-modal-root" />
            </div>
          </div>

          <div class="field" id="davs-modal-s3-container">
            <div class="field" id="davs-modal-endpoint-container">
              <label class="label">S3 backend endpoint</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-endpoint" />
              </div>
            </div>

            <div class="field" id="davs-modal-region-container">
              <label class="label">S3 backend region</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-region" />
              </div>
            </div>

            <div class="field" id="davs-modal-bucket-container">
              <label class="label">S3 backend bucket</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-bucket" />
              </div>
            </div>

            <div class="field" id="davs-modal-accesskeyid-container">
              <label class="label">S3 backend access key id</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-accesskeyid" />
              </div>
            </div>

            <div class="field" id="davs-modal-secretaccesskey-container">
              <label class="label">S3 backend secret access key</label>
              <div class="control">
                <input class="input" type="text" id="davs-modal-secretaccesskey" />
              </div>
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
  user = await Auth.GetUser();
  if (user !== undefined) {
    if (user.isAdmin) document.getElementById("davs-new").classList.toggle("is-hidden");
    registerModalFields();
    await firstShowDavs();
  }
}

function davTemplate(dav) {
  cleanDav(dav);
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
              ${user.isAdmin ? '<a class="dropdown-item" id="davs-dav-edit-' + dav.id + '"><i class="fas fa-edit"></i><strong> Edit</strong></a>' : ""}
              ${user.isAdmin ? '<a class="dropdown-item has-text-danger" id="davs-dav-delete-' + dav.id + '"><i class="fas fa-trash-alt"></i><strong> Delete</strong></a>' : ""}
              <hr class="dropdown-divider" />
              <div class="dropdown-item">
                <p>
                  <progress class="progress is-${GetColor(du)} is-small" value="${dav.usedgb}" max="${dav.totalgb}"></progress>${dav.usedgb !== undefined ? free + " GB free" : ""}
                </p>
                <hr class="dropdown-divider" />
                <p><strong>${dav.host}</strong></p>
                <p>serves ${dav.iss3 ? `${dav.bucket} bucket` : `${dav.root} directory`}, with ${dav.writable ? "read/write" : "read only"} access</p>
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

function displayDavs(inDavs) {
  const markup = inDavs
    .map((dav) => {
      if (user.isAdmin || !dav.secured || dav.roles.some((r) => user.memberOf.includes(r))) {
        return davTemplate(dav);
      }
    })
    .join("");
  document.getElementById("davs-list").innerHTML = markup;
  inDavs.map((dav) => {
    if (user.isAdmin) {
      document.getElementById(`davs-dav-edit-${dav.id}`).addEventListener("click", function () {
        editDav(dav);
      });
      document.getElementById(`davs-dav-delete-${dav.id}`).addEventListener("click", function () {
        new Delete(() => {
          deleteDav(dav);
        });
      });
    }
    if (user.isAdmin || !dav.secured || dav.roles.some((r) => user.memberOf.includes(r))) {
      document.getElementById(`davs-dav-open-${dav.id}`).addEventListener("click", function () {
        openExplorerModal(dav.host, dav.writable, dav.passphrase != null && dav.passphrase !== "");
      });
    }
  });
}

async function firstShowDavs() {
  try {
    const response = await fetch("/api/common/davs", {
      method: "GET",
      headers: new Headers({
        "XSRF-Token": user.xsrftoken,
      }),
    });
    if (response.status !== 200) {
      throw new Error(`Davs could not be fetched (status ${response.status})`);
    }
    davs = await response.json();
    displayDavs(davs);
  } catch (e) {
    HandleError(e);
  }
}

async function deleteDav(dav) {
  try {
    const response = await fetch("/api/admin/davs/" + dav.id, {
      method: "delete",
      headers: new Headers({
        "XSRF-Token": user.xsrftoken,
      }),
    });
    if (response.status !== 200) {
      throw new Error(`Dav could not be deleted (status ${response.status})`);
    }
    reloadDavsOnServer();
    document.getElementById(`davs-dav-${dav.id}`).remove();
  } catch (e) {
    HandleError(e);
  }
}

function registerModalFields() {
  id_field = document.getElementById("davs-modal-id");
  name_field = document.getElementById("davs-modal-name");
  icon_field = document.getElementById("davs-modal-icon");
  color_field = document.getElementById("davs-modal-color");
  host_field = document.getElementById("davs-modal-host");
  writable_field = document.getElementById("davs-modal-writable");
  root_field = document.getElementById("davs-modal-root");
  root_container = document.getElementById("davs-modal-root-container");
  secured_field = document.getElementById("davs-modal-secured");
  roles_field = document.getElementById("davs-modal-roles");
  roles_container = document.getElementById("davs-modal-roles-container");
  passphrase_field = document.getElementById("davs-modal-passphrase");
  iss3_field = document.getElementById("davs-modal-iss3");
  s3_container = document.getElementById("davs-modal-s3-container");
  endpoint_field = document.getElementById("davs-modal-endpoint");
  region_field = document.getElementById("davs-modal-region");
  bucket_field = document.getElementById("davs-modal-bucket");
  accesskeyid_field = document.getElementById("davs-modal-accesskeyid");
  secretaccesskey_field = document.getElementById("davs-modal-secretaccesskey");
  document.getElementById(`davs-modal-close`).addEventListener("click", function () {
    toggleModal();
  });
  document.getElementById(`davs-modal-cancel`).addEventListener("click", function () {
    toggleModal();
  });
  document.getElementById(`davs-modal-save`).addEventListener("click", function () {
    postDav();
  });
  document.getElementById(`davs-new`).addEventListener("click", function () {
    newDav();
  });
  icon_field.addEventListener("click", function () {
    pickIcon();
  });
  secured_field.addEventListener("click", function () {
    toggleRoles();
  });
  iss3_field.addEventListener("click", function () {
    toggleS3();
  });
  document.getElementById(`davs-modal-passphrase-generate`).addEventListener("click", function () {
    passphrase_field.value = RandomString(48);
  });
  document.getElementById(`davs-modal-passphrase-remove`).addEventListener("click", function () {
    passphrase_field.value = "";
  });
}

async function editDav(dav) {
  cleanDav(dav);
  id_field.value = dav.id;
  name_field.value = dav.name;
  icon_field.value = dav.icon;
  color_field.value = dav.color;
  host_field.value = dav.host;
  writable_field.checked = dav.writable;
  root_field.value = dav.root;
  secured_field.checked = dav.secured;
  roles_field.value = dav.roles;
  passphrase_field.value = dav.passphrase;
  iss3_field.checked = dav.iss3;
  endpoint_field.value = dav.endpoint;
  region_field.value = dav.region;
  bucket_field.value = dav.bucket;
  accesskeyid_field.value = dav.accesskeyid;
  secretaccesskey_field.value = dav.secretaccesskey;
  toggleModal();
}

function cleanDav(dav) {
  let props = ["writable", "name", "roles", "passphrase", "endpoint", "region", "bucket", "accesskeyid", "secretaccesskey"];
  for (const prop of props) {
    dav[prop] = dav[prop] === undefined ? "" : dav[prop];
  }
  dav.icon = dav.icon === undefined ? "file" : dav.icon;
}

async function newDav() {
  let maxid = 0;
  davs.map(function (dav) {
    if (dav.id > maxid) maxid = dav.id;
  });
  id_field.value = maxid + 1;
  name_field.value = "";
  icon_field.value = "file";
  color_field.value = "#000000";
  host_field.value = `new_dav_service.${location.hostname}`;
  writable_field.checked = false;
  root_field.value = "";
  secured_field.checked = false;
  roles_field.value = "";
  passphrase_field.value = "";
  iss3_field.checked = false;
  endpoint_field.value = "";
  region_field.value = "";
  bucket_field.value = "";
  accesskeyid_field.value = "";
  secretaccesskey_field.value = "";
  toggleModal();
}

async function postDav() {
  try {
    const response = await fetch("/api/admin/davs/", {
      method: "post",
      headers: new Headers({
        "XSRF-Token": user.xsrftoken,
      }),
      body: JSON.stringify({
        id: parseInt(id_field.value),
        name: name_field.value,
        icon: icon_field.value,
        color: color_field.value,
        host: host_field.value,
        writable: writable_field.checked,
        root: root_field.value,
        secured: secured_field.checked,
        roles: secured_field.checked ? roles_field.value.split(",") : "",
        passphrase: passphrase_field.value,
        iss3: iss3_field.checked,
        endpoint: endpoint_field.value,
        region: region_field.value,
        bucket: bucket_field.value,
        accesskeyid: accesskeyid_field.value,
        secretaccesskey: secretaccesskey_field.value,
      }),
    });
    if (response.status !== 200) {
      throw new Error(`Davs could not be updated (status ${response.status})`);
    }
    davs = await response.json();
    await displayDavs(davs);
    await reloadDavsOnServer();
  } catch (e) {
    HandleError(e);
  }
  toggleModal();
}

async function reloadDavsOnServer() {
  try {
    const response = await fetch("/api/admin/reload", {
      method: "GET",
      headers: new Headers({
        "XSRF-Token": user.xsrftoken,
      }),
    });
    if (response.status !== 200) {
      throw new Error(`Dav could not be reloaded (status ${response.status})`);
    }
  } catch (e) {
    HandleError(e);
  }
}

async function toggleModal() {
  toggleRoles();
  toggleS3();
  updateIcon();
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

function toggleRoles() {
  if (secured_field.checked) {
    roles_container.style.display = "block";
  } else {
    roles_container.style.display = "none";
  }
}

function toggleS3() {
  if (iss3_field.checked) {
    s3_container.style.display = "block";
    root_container.style.display = "none";
  } else {
    s3_container.style.display = "none";
    root_container.style.display = "block";
  }
}

function updateIcon() {
  document.getElementById("davs-modal-icon-preview").setAttribute("class", "fas fa-" + icon_field.value);
}

async function pickIcon() {
  const iconsTemplate =
    '<div class="buttons">' +
    Icons.map(
      (icon) => /* HTML */ `
        <button class="button${icon_field.value == icon ? " is-primary" : ""}" id="davs-icon-modal-list-${icon}">
          <span class="icon">
            <i class="fas fa-${icon}"></i>
          </span>
        </button>
      `
    ).join("") +
    "</div>";
  document.getElementById("davs-icons-modal-list").innerHTML = iconsTemplate;
  Icons.map((icon) =>
    document.getElementById(`davs-icon-modal-list-${icon}`).addEventListener("click", function () {
      icon_field.value = icon;
      updateIcon();
      document.getElementById("davs-icons-modal").classList.toggle("is-active");
    })
  );
  document.getElementById("davs-icons-modal").classList.toggle("is-active");
}

function openExplorerModal(hostname, readwrite, encrypted) {
  const modal = document.getElementById("davs-explorer-modal");
  const card = document.getElementById("davs-explorer-modal-card");
  const explorer = new Explorer(hostname);
  explorer.mount("davs-explorer-modal-card", readwrite, encrypted);
  modal.classList.add("is-active");
  AnimateCSS(modal, "fadeIn");
  AnimateCSS(card, "zoomIn");
}
