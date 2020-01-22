// Imports
import * as Auth from "/services/auth/auth.js";
import * as brand from "/assets/brand/brand.js";

// local variables
let user;
let menu;

export function mount(mountpoint) {
  const where = document.getElementById(mountpoint);
  window.document.title = brand.windowTitle;
  where.innerHTML = /* HTML */ `
    <div class="navbar-brand">
      <a class="navbar-item is-size-4" href="/"><img src="assets/brand/logo.svg" alt="logo" />${brand.navTitle}</a>
      <a role="button" id="navbar-burger" class="navbar-burger burger" aria-label="menu" aria-expanded="false">
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </a>
    </div>
    <div id="navbar-menu" class="navbar-menu"></div>
  `;
  // Hamburger menu
  const burger = document.getElementById("navbar-burger");
  menu = document.getElementById("navbar-menu");
  burger.addEventListener("click", () => {
    burger.classList.toggle("is-active");
    menu.classList.toggle("is-active");
  });
  menu.addEventListener("click", () => {
    if (burger.classList.contains("is-active")) {
      burger.classList.toggle("is-active");
      menu.classList.toggle("is-active");
    }
  });
  CreateMenu();
}

export async function CreateMenu() {
  user = await Auth.GetUser();
  menu.innerHTML = `
        <div class="navbar-start">
          ${
            user === undefined
              ? ``
              : /* HTML */ `
                  <a class="navbar-item" href="#apps"> Apps </a>
                  <a class="navbar-item" href="#davs"> Webdavs </a>
                  ${user.isAdmin
                    ? /* HTML */ `
                        <a class="navbar-item" href="#users"> Users </a>
                      `
                    : ""}
                `
          }
        </div>
        <div class="navbar-end">
          ${
            user === undefined
              ? /* HTML */ `
                  <a class="navbar-item" href="#login"> Log in </a>
                `
              : /* HTML */ `
                  <a class="navbar-item" href="/Logout"> Log out </a>
                `
          }
        </div>
     `;
}
