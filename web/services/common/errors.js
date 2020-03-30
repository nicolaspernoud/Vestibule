// Imports
import * as Messages from "/services/messages/messages.js";
import { DeleteUser } from "/services/auth/auth.js";

export function HandleError(error) {
  if (error.message.includes("401")) {
    Messages.Show("is-warning", "Authentication needed, please log in.");
    DeleteUser();
    location.href = "#login";
  } else {
    Messages.Show("is-warning", error.message);
    console.error(error.message);
  }
}
