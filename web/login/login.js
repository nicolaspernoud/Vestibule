// Imports
import * as Messages from "/messages/messages.js";

// DOM elements
let mountpoint;
let login_field;
let password_field;
let login_inmemory;
let login_oauth2;

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
        <a id="login-oauth2" class="card-footer-item">Login with OAuth2</a>
      </footer>
    </div>
  `;
  registerModalFields();
}

function registerModalFields() {
  login_field = document.getElementById("login-login");
  password_field = document.getElementById("login-password");
  login_inmemory = document.getElementById("login-inmemory");
  login_oauth2 = document.getElementById("login-oauth2");
  login_inmemory.addEventListener("click", function() {
    doLogin();
  });
  login_oauth2.addEventListener("click", function() {
    window.location.href = "/OAuth2Login";
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
    window.location.href = "/";
  } catch (e) {
    Messages.Show("is-warning", e);
  }
}
