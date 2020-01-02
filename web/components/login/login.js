// Imports
import * as Messages from "/services/messages/messages.js";

// DOM elements
let mountpoint;
let login_field;
let password_field;
let login_inmemory;

export function mount(where) {
  mountpoint = where;
  document.getElementById(mountpoint).innerHTML = /* HTML */ `
    <div class="card">
      <div class="card-content">
        <div class="field">
          <p class="control has-icons-left has-icons-right">
            <input id="login-login" class="input" type="text" placeholder="Login" />
            <span class="icon is-small is-left">
              <i class="fas fa-user"></i>
            </span>
          </p>
        </div>
        <div class="field">
          <p class="control has-icons-left">
            <input id="login-password" class="input" type="password" placeholder="Password" />
            <span class="icon is-small is-left">
              <i class="fas fa-lock"></i>
            </span>
          </p>
        </div>
      </div>
      <footer class="card-footer">
        <a id="login-inmemory" class="card-footer-item">Login</a>
        <a id="login-oauth2" class="card-footer-item" href="/OAuth2Login">Login with OAuth2</a>
      </footer>
    </div>
  `;
  registerModalFields();
}

function registerModalFields() {
  login_field = document.getElementById("login-login");
  password_field = document.getElementById("login-password");
  password_field.addEventListener("keyup", function(event) {
    // Number 13 is the "Enter" key on the keyboard
    if (event.keyCode === 13) {
      doLogin();
    }
  });
  login_inmemory = document.getElementById("login-inmemory");
  login_inmemory.addEventListener("click", function() {
    doLogin();
  });
}

async function doLogin() {
  try {
    const response = await fetch("/Login", {
      method: "POST",
      body: JSON.stringify({
        login: login_field.value,
        password: password_field.value
      })
    });
    if (response.status !== 200) {
      throw new Error(`Login error (status ${response.status})`);
    }
    location.hash = "#apps";
  } catch (e) {
    Messages.Show("is-warning", e.message);
    console.error(e);
  }
}
