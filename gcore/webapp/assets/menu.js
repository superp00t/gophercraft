window.updateMenuSize = () => {
  var wid = document.querySelector(".nav-bar").offsetWidth;

  document.querySelectorAll(".card").forEach((card) => {
    if ([...card.classList].includes("nav-bar"))
      return;
    var newWidth = wid + "px";
    card.style.width = wid + "px";
  });
}