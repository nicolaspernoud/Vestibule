// Imports
import { AnimateCSS } from "/services/common/common.js";

export class Delete {
  constructor(okFunction) {
    let deleteModal = document.createElement("div");
    deleteModal.classList.add("modal", "animate__animated", "animate__fadeIn", "is-active");
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
    const deleteOK = deleteModal.querySelector("#delete-ok");
    const deleteCancel = deleteModal.querySelector("#delete-cancel");
    const toggleButtons = () => {
      deleteOK.classList.toggle("is-loading");
      deleteOK.disabled = !deleteOK.disabled;
      deleteCancel.disabled = !deleteCancel.disabled;
    };
    deleteOK.addEventListener("click", async () => {
      toggleButtons();
      await okFunction();
      toggleButtons();
      await AnimateCSS(deleteModal, "fadeOut");
      deleteModal.parentNode.removeChild(deleteModal);
    });
    deleteCancel.addEventListener("click", async () => {
      await AnimateCSS(deleteModal, "fadeOut");
      deleteModal.parentNode.removeChild(deleteModal);
    });
    document.body.appendChild(deleteModal);
  }
}
