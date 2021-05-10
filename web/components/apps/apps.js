// Imports
import { Icons } from "/services/common/icons.js";
import { AnimateCSS } from "/services/common/common.js";
import { Delete } from "/services/common/delete.js";
import { HandleError } from "/services/common/errors.js";

export async function mount(where, user) {
  const appsComponent = new Apps(user);
  await appsComponent.mount(where);
}

class Apps {
  constructor(user) {
    this.user = user;
  }

  // DOM elements
  id_field;
  name_field;
  icon_field;
  color_field;
  host_field;
  isproxy_field;
  forwardto_field;
  serve_field;
  secured_field;
  login_field;
  password_field;
  openpath_field;
  roles_field;
  forwardto_container;
  serve_container;
  roles_container;
  securityheaders_field;
  cacheduration_field;
  cachepattern_field;

  // local variables
  apps;
  user;

  async mount(mountpoint) {
    document.getElementById(mountpoint).innerHTML = /* HTML */ `
      <div id="apps-list" class="flex-container"></div>
      <button id="apps-new" class="button is-primary is-hidden">
        <span class="icon is-small">
          <i class="fas fa-plus"></i>
        </span>
      </button>
      <div class="modal" id="apps-modal">
        <div class="modal-background"></div>
        <div class="modal-card" id="apps-modal-card">
          <header class="modal-card-head">
            <p class="modal-card-title">Add/Edit app</p>
            <button class="delete" aria-label="close" id="apps-modal-close"></button>
          </header>
          <section class="modal-card-body">
            <div class="field">
              <label class="label">Id (changing the id will overwrite an app with the same id, editing an app and changing for an unused id will clone the app)</label>
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
              <label class="label">Color</label>
              <div class="control">
                <input class="input" type="color" id="apps-modal-color" />
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
              <label class="label">Forward To (default protocol is http://, specify otherwise, example: https://localhost:4443)</label>
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
            <div class="field">
              <label class="label">Path on iframe opening</label>
              <div class="control">
                <input class="input" type="text" id="apps-modal-openpath" />
              </div>
            </div>
            <div class="field">
              <div class="control">
                <label class="label"><input id="apps-modal-securityheaders" type="checkbox" />Inject security headers (CSP, STS, etc.)</label>
              </div>
            </div>
            <div class="field">
              <label class="label">Cache GET requests for url path matching<br />(separated with commas - * replaces anything - empty: no cache - example: /api/*,*/admin)</label>
              <div class="control">
                <input class="input" type="text" id="apps-modal-cachepattern" />
              </div>
            </div>
            <div class="field">
              <label class="label">Server cache GET requests for seconds (0/empty to disable)</label>
              <div class="control">
                <input class="input" type="number" id="apps-modal-cacheduration" />
              </div>
            </div>
            <br />
          </section>
          <footer class="modal-card-foot">
            <button id="apps-modal-save" class="button is-success">Save changes</button>
            <button id="apps-modal-cancel" class="button">Cancel</button>
          </footer>
        </div>
      </div>

      <div class="modal animate__animated animate__zoomIn" id="apps-icons-modal">
        <div class="modal-card">
          <section id="apps-icons-modal-list" class="modal-card-body"></section>
        </div>
      </div>
    `;
    if (this.user.isAdmin) document.getElementById("apps-new").classList.toggle("is-hidden");
    this.registerModalFields();
    await this.firstShowApps();
  }

  appTemplate(app) {
    this.cleanApp(app);
    return /* HTML */ `
      <div id="apps-app-${app.id}" class="card icon-card">
        <div class="card-content has-text-centered" id="apps-app-open-${app.id}">
          <span class="icon is-medium" style="color:${app.color};">
            <i class="fas fa-3x fa-${app.icon ? app.icon : "file"}"></i>
          </span>
        </div>
        <p class="has-text-centered"><strong>${app.name ? app.name : app.id}</strong></p>
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
                <a class="dropdown-item" href="https://${app.host}:${location.port}"><i class="fas fa-external-link-alt"></i><strong> Visit</strong></a>
                ${this.user.isAdmin ? '<a class="dropdown-item" id="apps-app-edit-' + app.id + '"><i class="fas fa-edit"></i><strong> Edit</strong></a>' : ""}
                ${this.user.isAdmin
                  ? '<a class="dropdown-item has-text-danger" id="apps-app-delete-' + app.id + '"><i class="fas fa-trash-alt"></i><strong> Delete</strong></a>'
                  : ""}
                <hr class="dropdown-divider" />
                <div class="dropdown-item">
                  <p><strong>${app.host}</strong></p>
                  <p>${app.isProxy ? "Proxies to <strong>" + app.forwardTo + "</strong>" : "Serves <strong>" + app.serve + "</strong>"}</p>
                  <p>${app.secured ? "Restricted access to users with roles <strong>" + app.roles + "</strong>" : "Unrestricted access"}</p>
                  ${app.login ? "<p>Automatically log in with basic auth as <strong>" + app.login + "</strong></p>" : ""}
                  ${app.cacheduration > 0 ? `<p>Server cache GET requests for <strong>${app.cachepattern}</strong> for <strong>${app.cacheduration}</strong> seconds</p>` : ""}
                  <p class="has-text-centered"><strong>-${app.id}-</strong></p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  displayApps() {
    const markup = this.apps
      .map((app) => {
        if (this.user.isAdmin || !app.secured || app.roles.some((r) => this.user.memberOf.includes(r))) {
          return this.appTemplate(app);
        }
      })
      .join("");
    document.getElementById("apps-list").innerHTML = markup;
    this.apps.map((app) => {
      if (this.user.isAdmin) {
        document.getElementById(`apps-app-edit-${app.id}`).addEventListener("click", () => {
          this.editApp(app);
        });
        document.getElementById(`apps-app-delete-${app.id}`).addEventListener("click", () => {
          new Delete(() => {
            this.deleteApp(app);
          }, app.name);
        });
      }
      if (this.user.isAdmin || !app.secured || app.roles.some((r) => this.user.memberOf.includes(r))) {
        document.getElementById(`apps-app-open-${app.id}`).addEventListener("click", () => {
          this.openWebview(app);
        });
      }
    });
  }

  async firstShowApps() {
    try {
      const response = await fetch("/api/common/apps", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Apps could not be fetched (status ${response.status})`);
      }
      this.apps = await response.json();
      this.displayApps();
    } catch (e) {
      HandleError(e);
    }
  }

  async deleteApp(app) {
    try {
      const response = await fetch("/api/admin/apps/" + app.id, {
        method: "delete",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`App could not be deleted (status ${response.status})`);
      }
      this.reloadAppsOnServer();
      document.getElementById(`apps-app-${app.id}`).remove();
    } catch (e) {
      HandleError(e);
    }
  }

  registerModalFields() {
    this.id_field = document.getElementById("apps-modal-id");
    this.name_field = document.getElementById("apps-modal-name");
    this.icon_field = document.getElementById("apps-modal-icon");
    this.color_field = document.getElementById("apps-modal-color");
    this.host_field = document.getElementById("apps-modal-host");
    this.isproxy_field = document.getElementById("apps-modal-isproxy");
    this.forwardto_field = document.getElementById("apps-modal-forwardto");
    this.serve_field = document.getElementById("apps-modal-serve");
    this.secured_field = document.getElementById("apps-modal-secured");
    this.roles_field = document.getElementById("apps-modal-roles");
    this.login_field = document.getElementById("apps-modal-login");
    this.password_field = document.getElementById("apps-modal-password");
    this.openpath_field = document.getElementById("apps-modal-openpath");
    this.forwardto_container = document.getElementById("apps-modal-forwardto-container");
    this.serve_container = document.getElementById("apps-modal-serve-container");
    this.roles_container = document.getElementById("apps-modal-roles-container");
    this.securityheaders_field = document.getElementById("apps-modal-securityheaders");
    this.cacheduration_field = document.getElementById("apps-modal-cacheduration");
    this.cachepattern_field = document.getElementById("apps-modal-cachepattern");
    document.getElementById(`apps-modal-close`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`apps-modal-cancel`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`apps-modal-save`).addEventListener("click", () => {
      this.postApp();
    });
    document.getElementById(`apps-new`).addEventListener("click", () => {
      this.newApp();
    });
    this.icon_field.addEventListener("click", () => {
      this.pickIcon();
    });
    this.isproxy_field.addEventListener("click", () => {
      this.toggleForwardServe();
    });
    this.secured_field.addEventListener("click", () => {
      this.toggleRoles();
    });
  }

  async editApp(app) {
    this.cleanApp(app);
    this.id_field.value = app.id;
    this.name_field.value = app.name;
    this.icon_field.value = app.icon;
    this.color_field.value = app.color;
    this.host_field.value = app.host;
    this.isproxy_field.checked = app.isProxy;
    this.forwardto_field.value = app.forwardTo;
    this.serve_field.value = app.serve;
    this.secured_field.checked = app.secured;
    this.roles_field.value = app.roles;
    this.login_field.value = app.login;
    this.password_field.value = app.password;
    this.openpath_field.value = app.openpath;
    this.securityheaders_field.checked = app.securityheaders;
    this.cacheduration_field.value = app.cacheduration;
    this.cachepattern_field.value = app.cachepattern;
    this.toggleModal();
  }

  cleanApp(app) {
    let props = ["name", "forwardTo", "serve", "roles", "login", "password", "openpath", "cacheduration", "cachepattern"];
    for (const prop of props) {
      app[prop] = app[prop] === undefined ? "" : app[prop];
    }
    app.icon = app.icon === undefined ? "file" : app.icon;
  }

  async newApp() {
    let maxid = 0;
    this.apps.map(function (app) {
      if (app.id > maxid) maxid = app.id;
    });
    this.id_field.value = maxid + 1;
    this.name_field.value = "";
    this.icon_field.value = "file";
    this.color_field.value = "#000000";
    this.host_field.value = `*.new_application.${location.hostname}`;
    this.isproxy_field.checked = false;
    this.forwardto_field.value = "";
    this.serve_field.value = "";
    this.secured_field.checked = false;
    this.roles_field.value = "";
    this.login_field.value = "";
    this.password_field.value = "";
    this.openpath_field.value = "";
    this.securityheaders_field.checked = false;
    this.cacheduration_field.value = 0;
    this.cachepattern_field.value = "";
    this.toggleModal();
  }

  async postApp() {
    try {
      let bdy = {
        id: parseInt(this.id_field.value),
        name: this.name_field.value,
        icon: this.icon_field.value,
        color: this.color_field.value,
        host: this.host_field.value,
        isProxy: this.isproxy_field.checked,
        forwardTo: this.forwardto_field.value,
        serve: this.serve_field.value,
        secured: this.secured_field.checked,
        roles: this.secured_field.checked ? this.roles_field.value.split(",") : "",
        login: this.login_field.value,
        password: this.password_field.value,
        openpath: this.openpath_field.value,
        securityheaders: this.securityheaders_field.checked,
      };
      const cachepattern = this.cachepattern_field.value.split(",");
      if (cachepattern.length > 1 || cachepattern[0] !== "") {
        bdy.cachepattern = cachepattern;
        bdy.cacheduration = this.cacheduration_field.value > 0 ? parseInt(this.cacheduration_field.value) : "";
      }
      const response = await fetch("/api/admin/apps/", {
        method: "post",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        body: JSON.stringify(bdy),
      });
      if (response.status !== 200) {
        throw new Error(`Apps could not be updated (status ${response.status})`);
      }
      this.apps = await response.json();
      this.displayApps();
      await this.reloadAppsOnServer();
    } catch (e) {
      HandleError(e);
    }
    this.toggleModal();
  }

  async reloadAppsOnServer() {
    try {
      const response = await fetch("/api/admin/reload", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`App could not be reloaded (status ${response.status})`);
      }
    } catch (e) {
      HandleError(e);
    }
  }

  async toggleModal() {
    this.toggleForwardServe();
    this.toggleRoles();
    this.updateIcon();
    const modal = document.getElementById("apps-modal");
    const card = document.getElementById("apps-modal-card");
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

  toggleForwardServe() {
    if (this.isproxy_field.checked) {
      this.forwardto_container.style.display = "block";
      this.serve_container.style.display = "none";
    } else {
      this.forwardto_container.style.display = "none";
      this.serve_container.style.display = "block";
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
    document.getElementById("apps-modal-icon-preview").setAttribute("class", "fas fa-" + this.icon_field.value);
  }

  async pickIcon() {
    const iconsTemplate =
      '<div class="buttons">' +
      Icons.map(
        (icon) => /* HTML */ `
          <button class="button${this.icon_field.value == icon ? " is-primary" : ""} is-medium" id="apps-icon-modal-list-${icon}">
            <span class="icon">
              <i class="fas fa-${icon}"></i>
            </span>
          </button>
        `
      ).join("") +
      "</div>";
    document.getElementById("apps-icons-modal-list").innerHTML = iconsTemplate;
    Icons.map((icon) =>
      document.getElementById(`apps-icon-modal-list-${icon}`).addEventListener("click", () => {
        this.icon_field.value = icon;
        this.updateIcon();
        document.getElementById("apps-icons-modal").classList.toggle("is-active");
      })
    );
    document.getElementById("apps-icons-modal").classList.toggle("is-active");
  }

  openWebview(app) {
    const url = `https://${app.host}:${location.port}${app.openpath}`;
    let webview = document.createElement("div");
    webview.classList.add("modal", "is-active");
    webview.innerHTML = /* HTML */ `
      <div class="modal-background animate__animated animate__fadeIn"></div>
      <div class="modal-card animate__animated animate__zoomIn">
        <header class="modal-card-head">
          <span class="icon mr-2"> <i class="fas fa-lg fa-${app.icon}" style="color: ${app.color};"></i> </span>
          <p class="modal-card-title">${app.name}</p>
          <button class="delete" aria-label="close" id="apps-webview-close"></button>
        </header>
        <section class="modal-card-body is-paddingless">
          <iframe src="${url}" style="height: 75vh;width: 100%;"></iframe>
        </section>
        <footer class="modal-card-foot"></footer>
      </div>
    `;
    webview.querySelector("#" + "apps-webview-close").addEventListener("click", async () => {
      AnimateCSS(webview.getElementsByClassName("modal-background")[0], "fadeOut");
      await AnimateCSS(webview.getElementsByClassName("modal-card")[0], "zoomOut");
      webview.parentNode.removeChild(webview);
    });
    document.body.appendChild(webview);
  }
}
