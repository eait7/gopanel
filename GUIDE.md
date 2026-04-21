# GoPanel Ecosystem: Complete User Guide

Welcome to the GoPanel Ecosystem! This is your definitive manual for operating your new lightweight, zero-bloat server infrastructure. GoPanel unifies three powerful Go-native services into a single UI, giving you everything you need to host applications, proxy traffic, and manage files without the heavy overhead of traditional control panels.

---

## 1. Initial Setup & First Login

When you execute the `install.sh` command, your ecosystem spins up on the following ports:

*   **GoPanel Dashboard**: `http://<your-ip>:9000`
*   **FileBrowser**: `http://<your-ip>:8090`
*   **Portainer**: `https://<your-ip>:9443`

### The First Things You MUST Do:
1.  **Log into GoPanel** (`admin` / `admin`) and familiarize yourself with the interface.
2.  **Log into FileBrowser** (`admin` / `Admin123456!`). Instantly navigate to _Settings > Profile Management_ and change your password.
3.  **Log into Portainer**. Because this is the first time you are booting it, Portainer will ask you to create a brand new Administrator username and password.

---

## 2. Navigating the GoPanel Dashboard

The central GoPanel dashboard handles top-level orchestration. 

### System Overview
The main screen provides up-to-the-second readings on your Server load, RAM usage, Uptime, and CPU stress by securely querying your system's `loadavg` and `/proc` files. 

### Domain Management (The Reverse Proxy)
Our domain manager talks directly to the Caddy API. You can add domains to your server and Caddy will **automatically provision SSL certificates (Let's Encrypt)** for them. 

When you click "Add Domain", you have two options:
1.  **Reverse Proxy**: If you have a Docker container running an app on Port `3000`, add your domain (e.g., `app.mywebsite.com`) and set the upstream to `localhost:3000`. Caddy instantly routes all secure traffic from your domain to that port!
2.  **File Server**: If you just want to host a static website (HTML/CSS), select File Server, and point the upstream to a local folder in your `storage` directory. 

### Settings & Outbound Email
Navigate to **Settings** to configure the system SMTP relay.
1.  Select your preferred email provider from the dropdown. 
2.  A provider like **Resend**, **Amazon SES**, or **Mailgun** ensures extreme deliverability.
3.  Enter your API Key as the password, and test the connection. This setting persists securely in `/data/settings.json` and is used for outbound system alerts or developer integrations.

---

## 3. Storage Management: FileBrowser

FileBrowser acts as your modern "Web FTP." Everything inside FileBrowser maps exactly to the `/opt/gopanel/storage` folder on your server machine.

*   **Uploading Code**: Just drag and drop your web application files.
*   **Sharing Files**: You can generate temporary expiring links to share direct files with clients.
*   **Editing Code**: Clicking on any text or code file opens a built-in Monaco (VS Code) syntax editor in your browser.

**Pro-Tip**: If you create a folder named `/srv/my_react_app` in FileBrowser and upload your static build files there, you can jump over to GoPanel and create a **File Server Domain** pointing directly to `/srv/my_react_app` to make it instantly live!

---

## 4. App Deployment: Portainer

Portainer is where the heavy lifting happens. It visualizing the raw Docker Socket API.

### How to Host a Complex Project (e.g. A Node.js App or Database):
1.  Click into **Local Environment**.
2.  Navigate to **Containers** -> **Add Container**.
3.  Type in the Docker image you want (e.g. `mysql:8.0` or `node:18`).
4.  **Network Ports**: Map your container's internal port to an unused host port (For example, map Host: `8080` to Container: `80`). 
5.  Click **Deploy the container**.
6.  *Final Step*: Jump back to the **GoPanel Dashboard -> Domains**, and proxy your domain to `localhost:8080`. Your app is now live, secured, and proxy-routed!

### Volumes
If your container generates data (like a database), make sure to map a Docker Volume in Portainer so that your data survives server reboots!

---

## 5. Security & Maintenance

*   **Zero-Downtime Reboots**: Caddy reads configuration straight from memory. Even if you reboot GoPanel or Portainer, Caddy will continue proxying traffic seamlessly.
*   **Updates**: Because everything operates via standard Docker compose structures, updating the system is as simple as running `cd /opt/gopanel && git pull && docker compose up -d --build`.

**Enjoy your unified, composable infrastructure!**
