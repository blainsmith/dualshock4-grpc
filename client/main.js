const hid = require('node-hid');
const grpc = require('grpc');

const pb = grpc.load(`${__dirname}/../pb/events.proto`).pb;
const eventsClient = new pb.Events('localhost:1313', grpc.credentials.createInsecure());
const controller = new hid.HID(1356, 1476);

const signal = eventsClient.signal();
signal.on('data', (signal) => {
    console.log('Signal received', signal.signal);
    process.exit(signal.signal);
});

const color = eventsClient.color();
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

const track = eventsClient.track(() => {});
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