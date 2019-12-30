window.updateRealmlist = () => {
    fetch("/v1/realmList")
    .then((data) => {
        return data.json();
    })
    .then((data) => {
        // Erase.
        var rl = document.querySelector(".realmlist");
        rl.innerHTML = "";

        data.listing.map((val) => {
            var el = document.createElement("div");
            el.classList.add("realmlisting");

            var lastUpdated = new Date(val.lastUpdated).getTime();
            var nowMs = new Date().getTime();
            var howOld = nowMs - lastUpdated;
            var online = howOld <= 12000;
            var onlineStr = online == true ? "online" : "offline";

            var headerLine = document.createElement("span");
            headerLine.classList.add("realmheader");

            var icon = document.createElement("img");
            icon.setAttribute("src", "/assets/server_" + onlineStr + ".png");
            headerLine.appendChild(icon);

            var p = document.createElement("p");
            p.innerText = val.name;
            p.classList.add("realmtitle");
            headerLine.appendChild(p);
            el.appendChild(headerLine);

            var desc = document.createElement("p");
            desc.classList.add("realmdescription");
            desc.innerHTML = val.description;
            el.appendChild(desc);

            rl.appendChild(el);
        });
    });
}

window.addEventListener("load", () => {
    updateRealmlist();

    window.setInterval(() => {
        updateRealmlist();
    }, 10000);
});
