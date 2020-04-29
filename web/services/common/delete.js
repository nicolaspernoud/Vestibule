// Imports
import { AnimateCSS } from "/services/common/common.js";

export class Delete {
  constructor(okFunction) {
    let deleteModal = document.createElement("div");
    deleteModal.classList.add("modal", "animated", "fadeIn", "faster", "is-active");
    deleteModal.innerHTML = /* HTML */ `
      <div class="modal-background"></div>
      <div class="modal-content">
        <div class="box" style="margin: 2rem;">
          <div class="field">
            <label class="label">Confirm</label>
          </div>
          <div class="field is-grouped">
            <div class="control">
              <button id="delete-ok" class="button is-danger">
                <span class="icon"><i class="fas fa-check"></i></span><span>Delete</span>
              </button>
            </div>
            <div class="control">
              <button id="delete-cancel" class="button">
                <span class="icon"><i class="fas fa-times-circle"></i></span><span>Cancel</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    `;
    deleteModal.querySelector("#" + "delete-ok").addEventListener("click", async () => {
      await okFunction();
      AnimateCSS(deleteModal, "fadeOut", function () {
        deleteModal.parentNode.removeChild(deleteModal);
      });
    });
    deleteModal.querySelector("#" + "delete-cancel").addEventListener("click", () => {
      AnimateCSS(deleteModal, "fadeOut", function () {
        deleteModal.parentNode.removeChild(deleteModal);
      });
    });
    document.body.appendChild(deleteModal);
  }
}
