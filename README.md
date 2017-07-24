<h1> Home Box </h1>

<h4>Contents:</h4>
<h6>1. System Requirements </h6>
<h6>2. Steps to build the containers </h6>
<h6>3. Steps to run the containers </h6>
<h6>4. Review </h6>


<h2>System Requirements:</h2>

1. One Hardisk (500Mb or 1Tb)
2. Raspberry Pi 3 or any linux machine capable or running Docker 17.05 - CE 
3. Basic I/O's

Mount the hardisk on raspberry pi at /media/pi/Dragon location. We need to setup a overlay network on which we will connect both the containers on


<h2>Steps to build the containers</h2>

1. HomeBox Container : Clone the code into local repository and go to backend and do "docker build -t homebox:latest ." Wola we have homebox container image. Verify it using "docker images".
	
2. Nginx Container: "docker pull tobi312/rpi-nginx"

 
<h2>Steps to run the containers </h2>

1. Create Net1 Overlay Network using following command:

docker network create -d overlay --subnet=5.5.0.0/16 --gateway=5.5.5.254 --ip-range=5.5.5.0/24 Net1

2. Run Nginx Container using following command: 

docker service create --name Nginx-1 --network Net1 --mount type=bind,source=/home/pi/homebox/default,destination=/etc/nginx/sites-available/default --mount type=bind,source=/home/pi/homebox/html,destination=/var/www/html --mount type=bind,source=/home/pi/homebox/ssl,destination=/etc/nginx/conf.d --mount type=bind,source=/home/pi/homebox/hosts,destination=/etc/hosts --publish 443:443/tcp tobi312/rpi-nginx
  
3. Run HboxServer Container using following command:

docker service create --name HBoxServer --network Net1 --mount type=bind,source=/media/pi/Dragon,destination=/media/pi/Dragon hboxserver:0.1
  
<h2>Review</h2> 
Browse to raspberry pi's IP using https protocol and wola you are done



  
