# homebox
Home Dropbox

Please follow below steps to run your Home Dropbox

This require Docker 17.05 installed

1. Run Nginx Container using following command: 

  docker service create --name Nginx-1 --network Net1 --mount type=bind,source=/home/pi/homebox/default,destination=/etc/nginx/sites-available/default 
  --mount type=bind,source=/home/pi/homebox/html,destination=/var/www/html --mount type=bind,source=/home/pi/homebox/ssl,destination=/etc/nginx/conf.d 
  --mount type=bind,source=/home/pi/homebox/hosts,destination=/etc/hosts --publish 443:443/tcp tobi312/rpi-nginx
  
2. Run HboxServer Container using following command:

  docker service create --name AfwServer --network Net1 --mount type=bind,source=/media/pi/Dragon,destination=/media/pi/Dragon hboxserver:0.1
  
  
To build the conatiner:

docker built -t hboxserver:0.1 .

