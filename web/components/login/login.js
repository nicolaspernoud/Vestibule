// Imports
import * as Navbar from "/components/navbar/navbar.js";
import { loginModes } from "/assets/brand/brand.js";
import * as Auth from "/services/auth/auth.js";
import { HandleError } from "/services/common/errors.js";

// DOM elements
let mountpoint;
let login_field;
let password_field;
let login_inmemory;
let login_icon;

export async function mount(where) {
  mountpoint = where;
  let user = await Auth.GetUser();
  if (user !== undefined) {
    document.getElementById(mountpoint).innerHTML = "";
    location.hash = "#";
  } else {
    Navbar.CreateMenu();
    document.getElementById(mountpoint).innerHTML = /* HTML */ `
      <div class="columns">
        <div class="column is-half is-offset-one-quarter">
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
              ${loginModes.inmemory
                ? /* HTML */ `
                    <a id="login-inmemory" class="card-footer-item">
                      <span class="icon" id="login-icon"><i class="fas fa-key"></i></span>Login
                    </a>
                  `
                : ""}
              ${loginModes.oauth2
                ? /* HTML */ `
                    <a id="login-oauth2" class="card-footer-item" href="/OAuth2Login">
                      <span class="icon"><i class="fab fa-keycdn"></i></span>Login with OAuth2
                    </a>
                  `
                : ""}
            </footer>
          </div>
        </div>
      </div>
    `;
    registerModalFields();
  }
}

function registerModalFields() {
  login_field = document.getElementById("login-login");
  password_field = document.getElementById("login-password");
  password_field.addEventListener("keyup", function (event) {
    // Number 13 is the "Enter" key on the keyboard
    if (event.keyCode === 13) {
      doLogin();
    }
  });
  login_inmemory = document.getElementById("login-inmemory");
  login_inmemory.addEventListener("click", function () {
    doLogin();
  });
  login_icon = document.getElementById("login-icon");
}

async function doLogin() {
  login_icon.classList.add("fa-pulse");
  try {
    const response = await fetch("/Login", {
      method: "POST",
      body: JSON.stringify({
        login: login_field.value,
        password: password_field.value,
      }),
    });
    if (response.status !== 200) {
      throw new Error(`Login error (status ${response.status})`);
    }
    await Auth.GetUser();
    // Redirect to original subdomain if login was displayed after an authentication error on the original subdomain
    const urlParams = new URLSearchParams(window.location.search);
    const redirectAfterLogin = urlParams.get("redirectAfterLogin");
    if (redirectAfterLogin != "" && redirectAfterLogin != null) {
      window.location.replace("https://" + redirectAfterLogin);
    }
    location.hash = "#davs";
    Navbar.CreateMenu();
  } catch (e) {
    HandleError(e);
    login_icon.classList.remove("fa-pulse");
  }
}
