var express = require('express');
var mraa = require('mraa');
var exec = require('child_process').exec;
var jsonfile = require('jsonfile');

var router = express.Router();

//Global variables
var userbutton;
var ram_array;
var cpu_array;
var disk_array;

//JSON files for client
var input_file = 'public/json_interface/gpio_interface_input.json';
var network_file = 'public/json_interface/network_interface.json';
var usb_file = 'public/json_interface/usb_interface.json';
var task_file = 'public/json_interface/task_interface.json';

//configure GPIO
var userled = new mraa.Gpio(13);
userled.mode(mraa.PIN_GPIO);
userled.dir(mraa.DIR_OUT);

var DQ0 = new mraa.Gpio(8);
DQ0.mode(mraa.PIN_GPIO);
DQ0.dir(mraa.DIR_OUT);

var DQ1 = new mraa.Gpio(7);
DQ1.mode(mraa.PIN_GPIO);
DQ1.dir(mraa.DIR_OUT);

var DI0 = new mraa.Gpio(12);
DI0.mode(mraa.PIN_GPIO);
DI0.dir(mraa.DIR_IN);

var DI1 = new mraa.Gpio(11);
DI1.mode(mraa.PIN_GPIO);
DI1.dir(mraa.DIR_IN);

var DI2 = new mraa.Gpio(10);
DI2.mode(mraa.PIN_GPIO);
DI2.dir(mraa.DIR_IN);

var DI3 = new mraa.Gpio(9);
DI3.mode(mraa.PIN_GPIO);
DI3.dir(mraa.DIR_IN);

var DI4 = new mraa.Gpio(4);
DI4.mode(mraa.PIN_GPIO);
DI4.dir(mraa.DIR_IN);

var AIU0 = new mraa.Aio(0);
var AIU1 = new mraa.Aio(2);
var AII0 = new mraa.Aio(1);
var AII1 = new mraa.Aio(3);

//10s loop
var timer = setInterval(function()
{
	//Execute system command that returns performance stats of the system
	//It returns 2 lines: the first contains the stats since the last boot 
	//the second contains the stats of the last 7 seconds
	//at the moment only the cpu time of the second line is used
	exec("vmstat 7 2", function(error, stdout, stderr){
		//safe console output in string
		var mystring = JSON.stringify(stdout);
		//safe string in array and filter empty spaces
		cpu_array = mystring.split(/[ ,]+/);
	});

	//Execute system command that returns the current stats of the RAM
	exec("free -m", function(error, stdout, stderr){
		//safe console output in string
		var mystring = JSON.stringify(stdout);
		//safe string in array and filter empty spaces
		ram_array = mystring.split(/[ ,]+/);
	});

	//Execute system command that returns the current usage of the filesystem
	exec("df -h", function(error, stdout, stderr){
		//safe console output in string
		var mystring = JSON.stringify(stdout);
		//Safe string in array and filter empty spaces
		disk_array = mystring.split(/[ ,]+/);
	});

	var myJSON = {"totalram":0,"usedram":0,"sharedram":0,"cache":0,"availableram":0,"totalswap":0,"usedswap":0,"freeswap":0,"freeram":0,"usertime":0,"systime":0,"unusedtime":0,"waitingtime":0,"virtualtime":0,"size":0,"used":0,"usedperc":0};


	//If array exists
	if(typeof ram_array !== "undefined")
	{
		//important informations from array to JSON
		myJSON.totalram = ram_array[7];
		myJSON.usedram = ram_array[8];
		myJSON.freeram = ram_array[9];
		myJSON.sharedram = ram_array[10];
		myJSON.cache = ram_array[11];
		myJSON.availableram = parseInt(ram_array[12]);
		myJSON.totalswap = ram_array[13];
		myJSON.usedswap = ram_array[14];
		myJSON.freeswap = parseInt(ram_array[15]);
	}

	//If array exists
	if(typeof cpu_array !== "undefined")
	{
		//important informations from array to JSON
		myJSON.usertime = cpu_array[52];
		myJSON.systime = cpu_array[53];
		myJSON.unusedtime = cpu_array[54];
		myJSON.waitingtime = cpu_array[55];
		myJSON.virtualtime = parseInt(cpu_array[56]);
    }

    //If array exists
    if(typeof disk_array !== "undefined")
    {
    	//important informations from array to JSON
    	myJSON.size = disk_array[7];
    	myJSON.used = disk_array[8];
    	myJSON.usedperc = disk_array[10];
    }

	//If JSON exists
	if(myJSON){
	  //write back to JSON file
	  jsonfile.writeFile(task_file, myJSON);
	}
}, 10000)


//1s loop 
var timer = setInterval(function()
{
	//Execute system command that returns state of userbutton
	exec("cat /sys/class/gpio/gpio63/value", function(error, stdout, stderr) 
	{
		//safe state of userbutton in global variable
		userbutton = parseInt(stdout);
	});

	var myJSON = {"userbutton":0,"DI0":0,"DI1":0,"DI2":0,"DI3":0,"DI4":0,"AIU0":0,"AIU1":0,"AII0":0,"AII1":0};

	//Fetch current state of inputs and write to JavaScriptObject
	myJSON.userbutton = userbutton;
	myJSON.DI0 = DI0.read();
	myJSON.DI1 = DI1.read();
	myJSON.DI2 = DI2.read();
	myJSON.DI3 = DI3.read();
	myJSON.DI4 = DI4.read();
	myJSON.AIU0 = AIU0.read();
	myJSON.AII0 = AII0.read();
	myJSON.AIU1 = AIU1.read();
	myJSON.AII1 = AII1.read();

	//If JSON exists
	if(myJSON){
	  //write back to JSON file
	  jsonfile.writeFile(input_file, myJSON);
	}
},1000);



//GET home page
router.get('/', function(req, res) {
	res.render('index');
});

//GET about page
router.get('/about', function(req, res, next){
	res.render('about');
});

//GET network page
router.get('/network', function(req, res, next){

	//Execute system cummand that returns current network config
	exec("ifconfig", function(error, stdout, stderr) 
	{
		//split console output in the parts: eth0, eth1, lo, wlan0 and safe in network_array
		var network_array = stdout.split("eth1");
		var myarray = network_array[1].split("lo");		
		network_array[1] = myarray[0];
		myarray = myarray[1].split("wlan0");
		network_array[2] = myarray[0];
		network_array[3] = myarray[1];

		var myJSON = {"eth0":"","eth1":"","wlan0":""};

		//array parts to JSON 
		myJSON.eth0 = network_array[0];
		myJSON.eth1 = "eth1 " + network_array[1];
		myJSON.lo = "lo " + network_array[2];
		myJSON.wlan0 = "wlan0 " + network_array[3];

		if(myJSON){
	  		//write back to JSON file
	  		jsonfile.writeFile(network_file, myJSON);
		}
	});

	res.render('network');
});

//GET usb page
router.get('/usb', function(req, res, next){

	//Execute System command that lists USB devices (intern and extern)
	exec("lsusb", function(error, stdout, stderr)
	{
		//Split console output in different devices 
		var usb_array = stdout.split("Bus", 10)

		var myJSON = {"usb0":0,"usb1":0,"usb2":0,"usb3":0,"usb4":0,"usb5":0,"usb6":0,"usb7":0,"usb8":0,"usb9":0};


		//array parts to JSON
		myJSON.usb0 = usb_array[0];
		myJSON.usb1 = "Bus " + usb_array[1];
		myJSON.usb2 = "Bus " + usb_array[2];
		myJSON.usb3 = "Bus " + usb_array[3];
		myJSON.usb4 = "Bus " + usb_array[4];
		myJSON.usb5 = "Bus " + usb_array[5];
		myJSON.usb6 = "Bus " + usb_array[6];	
		myJSON.usb7 = "Bus " + usb_array[7];
		myJSON.usb8 = "Bus " + usb_array[8];
		myJSON.usb9 = "Bus " + usb_array[9];

		if(myJSON){
	  		//write back to JSON file
	  		jsonfile.writeFile(usb_file, myJSON);
		}
	});

	res.render('usb');
});

//GET taskmanager page
router.get('/taskmanager', function(req, res, next){
	res.render('taskmanager');
});


//POST to fetch the JSON data from the client
router.post('/submitJSON', function(req, res, next) {

	//Save JSON from cient
	var myJSON = req.body;

	//Write commands from JSON to GPIO
	userled.write(parseInt(myJSON.userled));
	DQ0.write(parseInt(myJSON.DQ0));
	DQ1.write(parseInt(myJSON.DQ1));

	//send response ok
	res.sendStatus(200);

});

module.exports = router;

