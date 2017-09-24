const hid = require('node-hid'); // package for interfacing with Human Interface Devices
const grpc = require('grpc'); // package for grpc

// Parse the .proto definition
const pb = grpc.load(`${__dirname}/../pb/events.proto`).pb;
// Create an instance of the event client by connecting to the running server
const eventsClient = new pb.Events('localhost:1313', grpc.credentials.createInsecure());
// Create a new controller connection via HID
const controller = new hid.HID(1356, 1476);

// Create a signal client
const signal = eventsClient.signal();
// Start a listener so every time we receive data we log it and then process.exit with that signal
signal.on('data', (signal) => {
    console.log('Signal received', signal.signal);
    process.exit(signal.signal);
});

// Create a color client
const color = eventsClient.color();
// Start a listener so every time we get a color we write that color to the controller to change the color
color.on('data', (color) => {
    console.log('Color received', color);
    controller.write([
        0x05,
        0xff,
        0x04,
        0x00,
        0,
        0,
        color.Red,
        color.Green,
        color.Blue,
        0,
        0
    ]);
});

// Create a track client
const track = eventsClient.track(() => {});
// Start a listener so every time the controller sends data we stream that data to the server
controller.on('data', (data) => {
    track.write({
        timestamp: Date.now(),
        player: "rblgk",
        state: data
    });
});

setTimeout(() => {
    process.exit();
}, 60000);