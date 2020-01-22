openDocument();

async function openDocument() {
  const urlParams = new URLSearchParams(window.location.search);
  const file = urlParams.get("file");
  const token = encodeURIComponent(urlParams.get("token"));
  const user = urlParams.get("user");
  const mtime = urlParams.get("mtime");
  const fileName = file.split("/").pop();
  const fileExtension = file.split(".").pop();
  const key = (await digestMessage(fileName + mtime)).substring(0, 20);
  const config = {
    document: {
      fileType: fileExtension,
      key: key,
      title: fileName,
      url: `${file}?token=${token}`
    },
    editorConfig: {
      lang: "fr-FR",
      mode: `${fileExtension === "docx" || fileExtension === "xlsx" || fileExtension === "pptx" ? "edit" : "view"}`,
      callbackUrl: `{{.Hostname}}/onlyoffice/save?file=${file}&token=${token}`,
      customization: {
        autosave: false
      },
      user: {
        id: user,
        name: user
      }
    }
  };
  var docEditor = new DocsAPI.DocEditor("placeholder", config);
}

async function digestMessage(message) {
  const msgUint8 = new TextEncoder().encode(message);
  const hashBuffer = await crypto.subtle.digest("SHA-256", msgUint8);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, "0")).join("");
  return hashHex;
}
