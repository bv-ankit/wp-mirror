# WordPress Update Server Setup

This project is a custom WordPress update server implemented in Go. It provides API endpoints for WordPress core, plugin, and theme updates, as well as file downloads.

## 1. Setting up the web server

### Installing Go and necessary dependencies

1. Install Go (version 1.16 or later):
   ```
   wget https://golang.org/dl/go1.16.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.16.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

2. Install project dependencies:
   ```
   go get github.com/gin-gonic/gin
   go get github.com/go-redis/redis/v8
   ```

### Setting up Redis with background disk persistence

1. Install Redis:
   ```
   sudo apt update
   sudo apt install redis-server
   ```

2. Configure Redis for persistence:
   Edit `/etc/redis/redis.conf`:
   ```
   appendonly yes
   appendfsync everysec
   ```

3. Restart Redis:
   ```
   sudo systemctl restart redis-server
   ```

## 2. Setting up disks and Nginx

### Configuring the network disk for storing zip files

1. Mount the network disk:
   ```
   sudo mkdir /mnt/wordpress-files
   sudo mount -t nfs nfs-server:/path/to/share /mnt/wordpress-files
   ```

2. Add the mount to `/etc/fstab` for persistence:
   ```
   nfs-server:/path/to/share /mnt/wordpress-files nfs defaults 0 0
   ```

### Setting up a reverse proxy (Nginx) to handle HTTPS and load balancing

1. Install Nginx:
   ```
   sudo apt install nginx
   ```

2. Configure Nginx:
   Create a new file `/etc/nginx/sites-available/wordpress-update-server`:

   ```nginx
   upstream wp_update_backend {
       server 127.0.0.1:8080;
       # Add more backend servers here for load balancing
   }

   server {
       listen 80;
       server_name your-domain.com;
       return 301 https://$server_name$request_uri;
   }

   server {
       listen 443 ssl;
       server_name your-domain.com;

       ssl_certificate /path/to/fullchain.pem;
       ssl_certificate_key /path/to/privkey.pem;

       location / {
           proxy_pass http://wp_update_backend;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

3. Enable the site and restart Nginx:
   ```
   sudo ln -s /etc/nginx/sites-available/wordpress-update-server /etc/nginx/sites-enabled/
   sudo nginx -t
   sudo systemctl restart nginx
   ```

## 3. Setting up background jobs

### Configuring systemd services for the main application and background jobs

1. Create a systemd service file for the main application:
   `/etc/systemd/system/wp-update-server.service`

   ```ini
   [Unit]
   Description=WordPress Update Server
   After=network.target

   [Service]
   ExecStart=/path/to/wp-update-server
   WorkingDirectory=/path/to/wp-update-server
   User=www-data
   Group=www-data
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

2. Create systemd service files for background jobs:
   `/etc/systemd/system/wp-update-checker.service`
   `/etc/systemd/system/wp-update-worker.service`

   (Use similar content as the main service file, adjusting the `ExecStart` path)

3. Enable and start the services:
   ```
   sudo systemctl enable wp-update-server wp-update-checker wp-update-worker
   sudo systemctl start wp-update-server wp-update-checker wp-update-worker
   ```

### Setting up log rotation and monitoring

1. Configure log rotation:
   Create `/etc/logrotate.d/wp-update-server`:

   ```
   /var/log/wp-update-server/*.log {
       daily
       missingok
       rotate 14
       compress
       delaycompress
       notifempty
       create 0640 www-data adm
       sharedscripts
       postrotate
           systemctl reload wp-update-server wp-update-checker wp-update-worker
       endscript
   }
   ```

2. Set up monitoring with Prometheus and Grafana (optional):
   - Install Prometheus and Grafana
   - Configure Prometheus to scrape metrics from your application
   - Set up Grafana dashboards to visualize the metrics

## Security and Optimization

1. Firewall configuration:
   ```
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. Regular system updates:
   ```
   sudo apt update && sudo apt upgrade
   ```

3. Optimize Nginx for high traffic:
   Edit `/etc/nginx/nginx.conf`:
   ```nginx
   worker_processes auto;
   worker_connections 1024;
   keepalive_timeout 65;
   gzip on;
   ```

4. Tune Redis for performance:
   Edit `/etc/redis/redis.conf`:
   ```
   maxmemory 1gb
   maxmemory-policy allkeys-lru
   ```

5. Implement rate limiting in Nginx to prevent abuse:
   Add to your Nginx server block:
   ```nginx
   limit_req_zone $binary_remote_addr zone=one:10m rate=1r/s;
   limit_req zone=one burst=5;
   ```

Remember to adjust paths, domain names, and other configuration details according to your specific setup. Regularly monitor your server's performance and adjust configurations as needed for optimal performance and security.
