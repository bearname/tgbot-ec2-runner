(async function () {
    await main();
}())

async function main() {
    let form = document.getElementById("form");
    let startDateElement = document.getElementById("startDate");
    let instanceIdElement = document.getElementById("instanceId");
    let instanceNameElement = document.getElementById("instanceName");
    let stopDateElement = document.getElementById("stopDate");
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        let name = instanceNameElement.value;
        let init = {
            method: "POST",
            body: JSON.stringify({
                "start": startDateElement.value,
                "stop": stopDateElement.value,
                "instanceName":  (name === null) ? "" : name,
                "instanceId": instanceIdElement.value,
            }),
        };

        const raw = await fetch('http://localhost:8090/addTask', init)
        let data = await raw.text();
        console.log(data);
    })
}