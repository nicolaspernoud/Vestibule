// Imports
import * as Messages from "/services/messages/messages.js";
import { AnimateCSS } from "/services/common/common.js";
import * as Auth from "/services/auth/auth.js";

export class Share {
  constructor(hostname, file) {
    this.hostname = hostname;
    this.file = file;
    this.url = `${hostname}${file.path}`;
    this.fullURL = `${location.protocol}//${hostname}${location.port !== "" ? ":" + location.port : ""}${file.path}`;
  }

  async show() {
    this.user = await Auth.GetUser();
    let shareModal = document.createElement("div");
    shareModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
    shareModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box" style="margin: 2rem;">
          <div class="field">
            <label class="label">Share with</label>
            <div class="control">
              <input id="explorer-share-for" class="input" type="text" />
            </div>
          </div>
          <div class="field">
            <label class="label">Days</label>
            <div class="control">
              <input id="explorer-share-howlong" class="input" type="number" value="7" />
            </div>
          </div>
          <div class="field is-grouped">
            <div class="control">
              <button id="explorer-share-ok" class="button is-success">
                <span class="icon is-small"><i class="fas fa-check"></i></span>
              </button>
            </div>
            <div class="control">
              <button id="explorer-share-cancel" class="button is-danger">
                <span class="icon is-small"><i class="fas fa-times-circle"></i></span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
    shareModal.querySelector("#" + "explorer-share-ok").addEventListener("click", async () => {
      try {
        const lifespan = parseInt(shareModal.querySelector("#" + "explorer-share-howlong").value);
        const response = await fetch(location.origin + "/api/common/Share", {
          method: "POST",
          headers: new Headers({
            "XSRF-Token": this.user.xsrftoken
          }),
          credentials: "include",
          body: JSON.stringify({
            sharedfor: shareModal.querySelector("#" + "explorer-share-for").value,
            lifespan: lifespan,
            url: this.url,
            readonly: true
          })
        });
        if (response.status !== 200) {
          throw new Error(`Share token could not be made (status ${response.status})`);
        }
        const shareToken = await response.text();
        // Create result modal
        let resultModal = document.createElement("div");
        resultModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
        resultModal.innerHTML = /* HTML */ `
          <div class="modal-background"></div>
          <div class="modal-content">
            <div class="box" style="margin: 2rem;">
              <div class="content is-small">
                <h1>This link will be available during ${lifespan} days</h1>
                <a href="${this.fullURL + "?token=" + encodeURIComponent(shareToken)}" class="button">
                  <span class="icon">
                    <i class="fas fa-link"></i>
                  </span>
                  <span>Download</span>
                </a>
              </div>
              <div class="field is-grouped">
                <div class="control">
                  <button id="explorer-result-close" class="button is-success">
                    <span class="icon is-small"><i class="fas fa-check"></i></span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        `;
        resultModal.querySelector("#" + "explorer-result-close").addEventListener("click", () => {
          AnimateCSS(resultModal, "fadeOut", function() {
            resultModal.parentNode.removeChild(resultModal);
          });
        });
        document.body.appendChild(resultModal);
      } catch (e) {
        Messages.Show("is-warning", e.message);
        console.error(e);
      }
      AnimateCSS(shareModal, "fadeOut", function() {
        shareModal.parentNode.removeChild(shareModal);
      });
    });
    shareModal.querySelector("#" + "explorer-share-cancel").addEventListener("click", () => {
      AnimateCSS(shareModal, "fadeOut", function() {
        shareModal.parentNode.removeChild(shareModal);
      });
    });
    document.body.appendChild(shareModal);
  }
}
