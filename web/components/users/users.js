// Imports
import { AnimateCSS, RandomString } from "/services/common/common.js";
import { Delete } from "/services/common/delete.js";
import { HandleError } from "/services/common/errors.js";

export async function mount(where, user) {
  const usersComponent = new Users(user);
  await usersComponent.mount(where);
}

class Users {
  constructor(user) {
    this.current_user = user;
  }

  // DOM elements
  id_field;
  login_field;
  password_field;
  name_field;
  surname_field;
  email_field;
  roles_field;

  // local variables
  users;

  async mount(mountpoint) {
    document.getElementById(mountpoint).innerHTML = /* HTML */ `
      <div class="table-container">
        <table class="table is-bordered is-narrow is-hoverable is-fullwidth">
          <thead>
            <tr class="is-selected">
              <th>Id</th>
              <th>Login</th>
              <th>Name</th>
              <th>Surname</th>
              <th>Email</th>
              <th>Roles</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody id="users"></tbody>
        </table>
      </div>
      <button id="users-new" class="button is-primary">
        <span class="icon is-small">
          <i class="fas fa-plus"></i>
        </span>
      </button>

      <div class="modal" id="users-modal">
        <div class="modal-background"></div>
        <div class="modal-card" id="users-modal-card">
          <header class="modal-card-head">
            <p class="modal-card-title">Add/Edit user</p>
            <button class="delete" aria-label="close" id="users-modal-close"></button>
          </header>
          <section class="modal-card-body">
            <div class="field">
              <label class="label">Id</label>
              <div class="control">
                <input class="input" type="number" id="users-modal-id" />
              </div>
            </div>
            <div class="field">
              <label class="label">Login</label>
              <div class="control">
                <input class="input" type="text" id="users-modal-login" />
              </div>
            </div>
            <label class="label">Password</label>
            <div class="field has-addons">
              <div class="control is-expanded">
                <input class="input" type="text" id="users-modal-password" placeholder="Leave empty to keep current password" />
              </div>
              <div class="control">
                <button id="users-modal-password-generate" class="button">
                  <span class="icon is-small">
                    <i class="fas fa-dice"></i>
                  </span>
                </button>
              </div>
            </div>
            <div class="field" id="users-modal-name-container">
              <label class="label">Name</label>
              <div class="control">
                <input class="input" type="text" id="users-modal-name" />
              </div>
            </div>
            <div class="field" id="users-modal-surname-container">
              <label class="label">Surname</label>
              <div class="control">
                <input class="input" type="text" id="users-modal-surname" />
              </div>
            </div>
            <div class="field" id="users-modal-email-container">
              <label class="label">Email</label>
              <div class="control">
                <input class="input" type="email" id="users-modal-email" />
              </div>
            </div>
            <div class="field" id="users-modal-roles-container">
              <label class="label">Roles (separated with commas)</label>
              <div class="control">
                <input class="input" type="text" id="users-modal-roles" />
              </div>
            </div>
          </section>
          <footer class="modal-card-foot">
            <button id="users-modal-save" class="button is-success">Save changes</button>
            <button id="users-modal-cancel" class="button">Cancel</button>
          </footer>
        </div>
      </div>
    `;
    this.registerModalFields();
    await this.firstShowUsers();
  }

  cleanUser(user) {
    let props = ["password", "name", "surname", "email", "memberOf"];
    for (const prop of props) {
      user[prop] = user[prop] === undefined ? "" : user[prop];
    }
  }

  userTemplate(user) {
    this.cleanUser(user);
    return /* HTML */ `
      <tr id="users-user-${user.id}">
        <th>${user.id}</th>
        <td>${user.login}</td>
        <td>${user.name}</td>
        <td>${user.surname}</td>
        <td>${user.email}</td>
        <td>${user.memberOf}</td>
        <td>
          <a id="users-user-edit-${user.id}" class="button is-link is-small">
            <span>Edit</span>
            <span class="icon is-small">
              <i class="fas fa-pen"></i>
            </span>
          </a>
          <a id="users-user-delete-${user.id}" class="button is-danger is-small">
            <span>Delete</span>
            <span class="icon is-small">
              <i class="fas fa-times"></i>
            </span>
          </a>
        </td>
      </tr>
    `;
  }

  displayUsers() {
    this.users.sort((a, b) => parseInt(a.id) - parseInt(b.id));
    const markup = this.users.map((user) => this.userTemplate(user)).join("");
    document.getElementById("users").innerHTML = markup;
    this.users.map((user) => {
      document.getElementById(`users-user-edit-${user.id}`).addEventListener("click", () => {
        this.editUser(user);
      });
      document.getElementById(`users-user-delete-${user.id}`).addEventListener("click", () => {
        new Delete(() => {
          this.deleteUser(user);
        }, user.login);
      });
    });
  }

  async firstShowUsers() {
    try {
      const response = await fetch("/api/admin/users/", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.current_user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Users could not be fetched (status ${response.status})`);
      }
      this.users = await response.json();
      this.displayUsers();
    } catch (e) {
      HandleError(e);
    }
  }

  async deleteUser(user) {
    try {
      const response = await fetch("/api/admin/users/" + user.id, {
        method: "delete",
        headers: new Headers({
          "XSRF-Token": this.current_user.xsrftoken,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`User could not be deleted (status ${response.status})`);
      }
      document.getElementById(`users-user-${user.id}`).remove();
    } catch (e) {
      HandleError(e);
    }
  }

  registerModalFields() {
    this.id_field = document.getElementById("users-modal-id");
    this.login_field = document.getElementById("users-modal-login");
    this.password_field = document.getElementById("users-modal-password");
    this.name_field = document.getElementById("users-modal-name");
    this.surname_field = document.getElementById("users-modal-surname");
    this.email_field = document.getElementById("users-modal-email");
    this.roles_field = document.getElementById("users-modal-roles");
    document.getElementById(`users-modal-close`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`users-modal-cancel`).addEventListener("click", () => {
      this.toggleModal();
    });
    document.getElementById(`users-modal-save`).addEventListener("click", () => {
      this.postUser();
    });
    document.getElementById(`users-new`).addEventListener("click", () => {
      this.newUser();
    });
    document.getElementById(`users-modal-password-generate`).addEventListener("click", () => {
      this.password_field.value = RandomString(48);
    });
  }

  async editUser(user) {
    this.cleanUser(user);
    this.id_field.value = user.id;
    this.login_field.value = user.login;
    this.password_field.value = user.passwordHash !== "" ? "" : RandomString(48);
    this.name_field.value = user.name;
    this.surname_field.value = user.surname;
    this.email_field.value = user.email;
    this.roles_field.value = user.memberOf;
    this.toggleModal();
  }

  async newUser() {
    let maxid = 0;
    this.users.map(function (user) {
      if (parseInt(user.id) > maxid) maxid = user.id;
    });
    maxid++;
    this.id_field.value = maxid.toString();
    this.login_field.value = "";
    this.password_field.value = RandomString(48);
    this.name_field.value = "";
    this.surname_field.value = "";
    this.email_field.value = "";
    this.roles_field.value = "";
    this.toggleModal();
  }

  async postUser() {
    try {
      const response = await fetch("/api/admin/users/", {
        method: "POST",
        headers: new Headers({
          "XSRF-Token": this.current_user.xsrftoken,
        }),
        body: JSON.stringify({
          id: this.id_field.value,
          login: this.login_field.value,
          password: this.password_field.value,
          name: this.name_field.value,
          surname: this.surname_field.value,
          email: this.email_field.value,
          memberOf: this.roles_field.value.split(","),
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Users could not be updated (status ${response.status})`);
      }
      this.users = await response.json();
      this.displayUsers();
    } catch (e) {
      HandleError(e);
    }
    this.toggleModal();
  }

  async toggleModal() {
    const modal = document.getElementById("users-modal");
    const card = document.getElementById("users-modal-card");
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
}
