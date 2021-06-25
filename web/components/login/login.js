// Imports
import { loginModes } from "/assets/brand/brand.js";
import * as Auth from "/services/auth/auth.js";
import { HandleError } from "/services/common/errors.js";
import { IsEmpty } from "/services/common/common.js";

export class Login {
  constructor(user, navbar) {
    this.user = user;
    this.navbar = navbar;
  }

  // DOM elements
  login_field;
  password_field;
  login_inmemory;
  login_icon;

  async mount(mountpoint) {
    if (!IsEmpty(this.user)) {
      document.getElementById(mountpoint).innerHTML = "";
      location.hash = "#";
    } else {
      this.navbar.CreateMenu();
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
      this.registerModalFields();
    }
  }

  registerModalFields() {
    this.login_field = document.getElementById("login-login");
    this.password_field = document.getElementById("login-password");
    this.password_field.addEventListener("keyup", (event) => {
      // Number 13 is the "Enter" key on the keyboard
      if (event.keyCode === 13) {
        this.doLogin();
      }
    });
    this.login_inmemory = document.getElementById("login-inmemory");
    this.login_inmemory.addEventListener("click", () => {
      this.doLogin();
    });
    this.login_icon = document.getElementById("login-icon");
  }

  async doLogin() {
    this.login_icon.classList.add("fa-pulse");
    try {
      const response = await fetch("/Login", {
        method: "POST",
        body: JSON.stringify({
          login: this.login_field.value,
          password: this.password_field.value,
        }),
      });
      if (response.status !== 200) {
        throw new Error(`Login error (status ${response.status})`);
      }
      await Auth.GetUser();
      location.hash = "#davs";
      this.navbar.CreateMenu();
    } catch (e) {
      HandleError(e);
      this.login_icon.classList.remove("fa-pulse");
    }
  }
}
