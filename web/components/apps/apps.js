// Imports
import * as Messages from "/services/messages/messages.js";
import * as Auth from "/services/auth/auth.js";
import { Icons } from "./icons.js";
import { AnimateCSS } from "/services/common/common.js";

// DOM elements
let mountpoint;
let id_field;
let name_field;
let icon_field;
let host_field;
let isproxy_field;
let forwardto_field;
let serve_field;
let secured_field;
let login_field;
let password_field;
let roles_field;
let forwardto_container;
let serve_container;
let roles_container;

// local variables
let apps;
let user;

export async function mount(where) {
  mountpoint = where;
  document.getElementById(mountpoint).innerHTML = /* HTML */ `
    <div id="apps-list" class="columns is-multiline is-centered"></div>
    <button id="apps-new" class="button is-primary is-rounded is-hidden">+</button>

    <div class="modal" id="apps-modal">
      <div class="modal-background"></div>
      <div class="modal-card" id="apps-modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">Add/Edit app</p>
          <button class="delete" aria-label="close" id="apps-modal-close"></button>
        </header>
        <section class="modal-card-body">
          <div class="field">
            <label class="label">Id</label>
            <div class="control">
              <input class="input" type="number" id="apps-modal-id" />
            </div>
          </div>
          <div class="field">
            <label class="label">Name</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-name" />
            </div>
          </div>
          <div class="field">
            <label class="label">Icon</label>
            <div class="control has-icons-left">
              <input class="input" type="text" id="apps-modal-icon" />
              <span class="icon is-small is-left has-text-info">
                <i id="apps-modal-icon-preview" class="fas fa-file"></i>
              </span>
            </div>
          </div>
          <div class="field">
            <label class="label">Host (allow subdomains with "*." prefix)</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-host" />
            </div>
          </div>
          <div class="field">
            <div class="control">
              <label class="label"><input id="apps-modal-isproxy" type="checkbox" />App proxies to a server</label>
            </div>
          </div>
          <div class="field" id="apps-modal-forwardto-container">
            <label class="label">Forward To</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-forwardto" />
            </div>
          </div>
          <div class="field" id="apps-modal-serve-container">
            <label class="label">Serve</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-serve" />
            </div>
          </div>
          <div class="field">
            <div class="control">
              <label class="label"><input id="apps-modal-secured" type="checkbox" />Secure access to app</label>
            </div>
          </div>
          <div class="field" id="apps-modal-roles-container">
            <label class="label">Allow access to roles (separated with commas)</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-roles" />
            </div>
          </div>
          <div class="field">
            <label class="label">Injected Basic Auth login</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-login" />
            </div>
          </div>
          <div class="field">
            <label class="label">Injected Basic Auth password</label>
            <div class="control">
              <input class="input" type="text" id="apps-modal-password" />
            </div>
          </div>
        </section>
        <footer class="modal-card-foot">
          <button id="apps-modal-save" class="button is-success">Save changes</button>
          <button id="apps-modal-cancel" class="button">Cancel</button>
        </footer>
      </div>
    </div>

    <div class="modal animated zoomIn faster" id="apps-icons-modal">
      <div class="modal-card">
        <section id="apps-icons-modal-list" class="modal-card-body"></section>
      </div>
    </div>
  `;
  user = await Auth.GetUser();
  if (!(user === undefined) && user.isAdmin) {
    document.getElementById("apps-new").classList.toggle("is-hidden");
  }
  registerModalFields();
  await firstShowApps();
}

function appTemplate(app) {
  cleanApp(app);
  return /* HTML */ `
    <div id="apps-app-${app.id}" class="column is-two-thirds">
      <div class="card">
        <header class="card-header">
          <p class="card-header-title has-text-white">
            <span class="icon has-text-warning"><i class="fas fa-${app.icon ? app.icon : "file"}"></i></span>${app.name ? app.name : app.id} - ${app.host}
          </p>
        </header>
        <div class="card-content has-reduced-padding">
          <p>${app.isProxy ? "Proxies to <strong>" + app.forwardTo + "</strong>" : "Serves <strong>" + app.serve + "</strong>"}</p>
          <p>${app.secured ? "Restricted access to user with roles <strong>" + app.roles + "</strong>" : "Unrestricted access"}</p>
          ${app.login ? "<p>Automatically log in with basic auth as <strong>" + app.login + "</strong></p>" : ""}
        </div>
        <footer class="card-footer">
          <a class="card-footer-item" onclick="window.location.href = 'https://${app.host}:${location.port}'">Visit</a>
          ${user.isAdmin ? '<a class="card-footer-item" id="apps-app-edit-' + app.id + '">Edit</a>' : ""}
          ${user.isAdmin ? '<a class="card-footer-item" id="apps-app-delete-' + app.id + '">Delete</a>' : ""}
        </footer>
      </div>
    </div>
  `;
}

function displayApps(apps) {
  const markup = apps.map(app => appTemplate(app)).join("");
  document.getElementById("apps-list").innerHTML = markup;
  if (user.isAdmin) {
    apps.map(app => {
      document.getElementById(`apps-app-edit-${app.id}`).addEventListener("click", function() {
        editApp(app);
      });
      document.getElementById(`apps-app-delete-${app.id}`).addEventListener("click", function() {
        deleteApp(app);
      });
    });
  }
}

async function firstShowApps() {
  try {
    const response = await fetch("/api/common/apps", {
      method: "get"
    });
    if (response.status !== 200) {
      throw new Error(`Apps could not be fetched (status ${response.status})`);
    }
    apps = await response.json();
    displayApps(apps);
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}

async function deleteApp(app) {
  try {
    const response = await fetch("/api/admin/apps/" + app.id, {
      method: "delete"
    });
    if (response.status !== 200) {
      throw new Error(`App could not be deleted (status ${response.status})`);
    }
    reloadAppsOnServer();
    document.getElementById(`apps-app-${app.id}`).remove();
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}

function registerModalFields() {
  id_field = document.getElementById("apps-modal-id");
  name_field = document.getElementById("apps-modal-name");
  icon_field = document.getElementById("apps-modal-icon");
  host_field = document.getElementById("apps-modal-host");
  isproxy_field = document.getElementById("apps-modal-isproxy");
  forwardto_field = document.getElementById("apps-modal-forwardto");
  serve_field = document.getElementById("apps-modal-serve");
  secured_field = document.getElementById("apps-modal-secured");
  roles_field = document.getElementById("apps-modal-roles");
  login_field = document.getElementById("apps-modal-login");
  password_field = document.getElementById("apps-modal-password");
  forwardto_container = document.getElementById("apps-modal-forwardto-container");
  serve_container = document.getElementById("apps-modal-serve-container");
  roles_container = document.getElementById("apps-modal-roles-container");
  document.getElementById(`apps-modal-close`).addEventListener("click", function() {
    toggleModal();
  });
  document.getElementById(`apps-modal-cancel`).addEventListener("click", function() {
    toggleModal();
  });
  document.getElementById(`apps-modal-save`).addEventListener("click", function() {
    postApp();
  });
  document.getElementById(`apps-new`).addEventListener("click", function() {
    newApp();
  });
  icon_field.addEventListener("click", function() {
    pickIcon();
  });
  isproxy_field.addEventListener("click", function() {
    toggleForwardServe();
  });
  secured_field.addEventListener("click", function() {
    toggleRoles();
  });
}

async function editApp(app) {
  cleanApp(app);
  id_field.value = app.id;
  name_field.value = app.name;
  icon_field.value = app.icon;
  host_field.value = app.host;
  isproxy_field.checked = app.isProxy;
  forwardto_field.value = app.forwardTo;
  serve_field.value = app.serve;
  secured_field.checked = app.secured;
  roles_field.value = app.roles;
  login_field.value = app.login;
  password_field.value = app.password;
  toggleModal();
}

function cleanApp(app) {
  let props = ["name", "forwardTo", "serve", "roles", "login", "password"];
  for (const prop of props) {
    app[prop] = app[prop] === undefined ? "" : app[prop];
  }
  app.icon = app.icon === undefined ? "file" : app.icon;
}

async function newApp() {
  let maxid = 0;
  apps.map(function(app) {
    if (app.id > maxid) maxid = app.id;
  });
  id_field.value = maxid + 1;
  name_field.value = "";
  icon_field.value = "file";
  host_field.value = `*.new_application.${location.hostname}`;
  isproxy_field.checked = false;
  forwardto_field.value = "";
  serve_field.value = "";
  secured_field.checked = false;
  roles_field.value = "";
  login_field.value = "";
  password_field.value = "";
  toggleModal();
}

async function postApp() {
  try {
    const response = await fetch("/api/admin/apps/", {
      method: "post",
      body: JSON.stringify({
        id: parseInt(id_field.value),
        name: name_field.value,
        icon: icon_field.value,
        host: host_field.value,
        isProxy: isproxy_field.checked,
        forwardTo: forwardto_field.value,
        serve: serve_field.value,
        secured: secured_field.checked,
        roles: roles_field.value.split(","),
        login: login_field.value,
        password: password_field.value
      })
    });
    if (response.status !== 200) {
      throw new Error(`Apps could not be updated (status ${response.status})`);
    }
    apps = await response.json();
    await displayApps(apps);
    await reloadAppsOnServer(apps);
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
  toggleModal();
}

async function reloadAppsOnServer() {
  try {
    const response = await fetch("/api/admin/apps/reload", {
      method: "get"
    });
    if (response.status !== 200) {
      throw new Error(`App could not be reloaded (status ${response.status})`);
    }
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}

function toggleModal() {
  toggleForwardServe();
  toggleRoles();
  updateIcon();
  const modal = document.getElementById("apps-modal");
  const card = document.getElementById("apps-modal-card");
  if (modal.classList.contains("is-active")) {
    AnimateCSS(modal, "fadeOut");
    AnimateCSS(card, "zoomOut", function() {
      modal.classList.remove("is-active");
    });
  } else {
    modal.classList.add("is-active");
    AnimateCSS(modal, "fadeIn");
    AnimateCSS(card, "zoomIn");
  }
}

function toggleForwardServe() {
  if (isproxy_field.checked) {
    forwardto_container.style.display = "block";
    serve_container.style.display = "none";
  } else {
    forwardto_container.style.display = "none";
    serve_container.style.display = "block";
  }
}

function toggleRoles() {
  if (secured_field.checked) {
    roles_container.style.display = "block";
  } else {
    roles_container.style.display = "none";
  }
}

function updateIcon() {
  document.getElementById("apps-modal-icon-preview").setAttribute("class", "fas fa-" + icon_field.value);
}

async function pickIcon() {
  const iconsTemplate =
    '<div class="buttons">' +
    Icons.map(
      icon => /* HTML */ `
        <button class="button${icon_field.value == icon ? " is-primary" : ""}" id="apps-icon-modal-list-${icon}">
          <span class="icon">
            <i class="fas fa-${icon}"></i>
          </span>
        </button>
      `
    ).join("") +
    "</div>";
  document.getElementById("apps-icons-modal-list").innerHTML = iconsTemplate;
  Icons.map(icon =>
    document.getElementById(`apps-icon-modal-list-${icon}`).addEventListener("click", function() {
      icon_field.value = icon;
      updateIcon();
      document.getElementById("apps-icons-modal").classList.toggle("is-active");
    })
  );
  document.getElementById("apps-icons-modal").classList.toggle("is-active");
}
