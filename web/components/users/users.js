// Imports
import * as Messages from "/services/messages/messages.js";
import { AnimateCSS, RandomString } from "/services/common/common.js";
import * as Auth from "/services/auth/auth.js";
import { Delete } from "/services/common/delete.js";

// DOM elements
let mountpoint;
let id_field;
let login_field;
let password_field;
let passwordhash_field;
let name_field;
let surname_field;
let roles_field;

// local variables
let users;
let current_user;

export async function mount(where) {
  mountpoint = where;
  document.getElementById(mountpoint).innerHTML = /* HTML */ `
    <div class="table-container">
      <table class="table is-bordered is-narrow is-hoverable is-fullwidth">
        <thead>
          <tr class="is-selected">
            <th>Id</th>
            <th>Login</th>
            <th>Password</th>
            <th>Hash</th>
            <th>Name</th>
            <th>Surname</th>
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
            <label>Id</label>
            <div class="control">
              <input class="input" type="number" id="users-modal-id" />
            </div>
          </div>
          <div class="field">
            <label>Login</label>
            <div class="control">
              <input class="input" type="text" id="users-modal-login" />
            </div>
          </div>
          <div class="field" id="users-modal-password-container">
            <label>Password</label>
            <div class="control">
              <input class="input" type="text" id="users-modal-password" />
            </div>
          </div>
          <div class="field" id="users-modal-password-container">
            <label>Password Hash</label>
            <div class="control">
              <input class="input" type="text" id="users-modal-passwordhash" />
            </div>
          </div>
          <div class="field" id="users-modal-name-container">
            <label>Name</label>
            <div class="control">
              <input class="input" type="text" id="users-modal-name" />
            </div>
          </div>
          <div class="field" id="users-modal-surname-container">
            <label>Surname</label>
            <div class="control">
              <input class="input" type="text" id="users-modal-surname" />
            </div>
          </div>
          <div class="field" id="users-modal-roles-container">
            <label>Roles (separated with commas)</label>
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
  registerModalFields();
  current_user = await Auth.GetUser();
  await firstShowUsers();
}

function cleanUser(user) {
  let props = ["password", "name", "surname", "memberOf"];
  for (const prop of props) {
    user[prop] = user[prop] === undefined ? "" : user[prop];
  }
}

function userTemplate(user) {
  cleanUser(user);
  return /* HTML */ `
    <tr id="users-user-${user.id}">
      <th>${user.id}</th>
      <td>${user.login}</td>
      <td>${user.password}</td>
      <td>${user.passwordHash === undefined ? "" : "..."}</td>
      <td>${user.name}</td>
      <td>${user.surname}</td>
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

function displayUsers() {
  const markup = users.map(user => userTemplate(user)).join("");
  document.getElementById("users").innerHTML = markup;
  users.map(user => {
    document.getElementById(`users-user-edit-${user.id}`).addEventListener("click", function() {
      editUser(user);
    });
    document.getElementById(`users-user-delete-${user.id}`).addEventListener("click", function() {
      new Delete(() => {
        deleteUser(user);
      });
    });
  });
}

async function firstShowUsers() {
  try {
    const response = await fetch("/api/admin/users/", {
      method: "GET",
      headers: new Headers({
        "XSRF-Token": current_user.xsrftoken
      })
    });
    if (response.status !== 200) {
      throw new Error(`Users could not be fetched (status ${response.status})`);
    }
    users = await response.json();
    displayUsers();
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}

async function deleteUser(user) {
  try {
    const response = await fetch("/api/admin/users/" + user.id, {
      method: "delete",
      headers: new Headers({
        "XSRF-Token": current_user.xsrftoken
      })
    });
    if (response.status !== 200) {
      throw new Error(`User could not be deleted (status ${response.status})`);
    }
    document.getElementById(`users-user-${user.id}`).remove();
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}

function registerModalFields() {
  id_field = document.getElementById("users-modal-id");
  login_field = document.getElementById("users-modal-login");
  password_field = document.getElementById("users-modal-password");
  passwordhash_field = document.getElementById("users-modal-passwordhash");
  name_field = document.getElementById("users-modal-name");
  surname_field = document.getElementById("users-modal-surname");
  roles_field = document.getElementById("users-modal-roles");
  document.getElementById(`users-modal-close`).addEventListener("click", function() {
    toggleModal();
  });
  document.getElementById(`users-modal-cancel`).addEventListener("click", function() {
    toggleModal();
  });
  document.getElementById(`users-modal-save`).addEventListener("click", function() {
    postUser();
  });
  document.getElementById(`users-new`).addEventListener("click", function() {
    newUser();
  });
  password_field.addEventListener("click", function() {
    password_field.value = RandomString(48);
  });
}

async function editUser(user) {
  cleanUser(user);
  id_field.value = user.id;
  login_field.value = user.login;
  password_field.value = user.passwordHash !== "" ? "" : RandomString(48);
  passwordhash_field.value = user.passwordHash;
  name_field.value = user.name;
  surname_field.value = user.surname;
  roles_field.value = user.memberOf;
  toggleModal();
}

async function newUser() {
  let maxid = 0;
  users.map(function(user) {
    if (parseInt(user.id) > maxid) maxid = user.id;
  });
  maxid++;
  id_field.value = maxid.toString();
  login_field.value = "";
  password_field.value = RandomString(48);
  passwordhash_field.value = "";
  name_field.value = "";
  surname_field.value = "";
  roles_field.value = "";
  toggleModal();
}

async function postUser() {
  try {
    const response = await fetch("/api/admin/users/", {
      method: "POST",
      headers: new Headers({
        "XSRF-Token": current_user.xsrftoken
      }),
      body: JSON.stringify({
        id: id_field.value,
        login: login_field.value,
        password: password_field.value,
        passwordHash: passwordhash_field.value,
        name: name_field.value,
        surname: surname_field.value,
        memberOf: roles_field.value.split(",")
      })
    });
    if (response.status !== 200) {
      throw new Error(`Users could not be updated (status ${response.status})`);
    }
    users = await response.json();
    displayUsers();
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
  toggleModal();
}

function toggleModal() {
  const modal = document.getElementById("users-modal");
  const card = document.getElementById("users-modal-card");
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
