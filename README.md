# Rclone Mount Manager

A lightweight GUI tool to manage and monitor your Rclone mounts on Windows.  
Provides easy access to mount/unmount operations, status updates, and quick access to Rclone configuration.

---

## Features

- Mount and unmount Rclone remotes with a single click
- Live status updates for active mounts
- Searchable remote configuration list
- Quick access to Rclone configuration (`rclone config`)
- Status indicators:
  - 🔄 Mounting
  - ✅ Connected
  - 💤 Inactive

---

## Installation

1. Clone the repository:

```bash
git clone https://github.com/keshon/rclone-mount-manager.git
cd rclone-mount-manager
````

2. Make sure your Rclone executable is in the project folder:

```
./rclone/rclone.exe
```

3. Build the GUI (requires Go 1.24+):

```bash
go build -o rclone-mount.exe ./cmd/rclone-mount
```

4. Run the application:

```bash
./rclone-mount.exe
```

---

## Usage

* Select a remote from the dropdown list
* Click **Mount Remote** to mount
* Click **Unmount Remote** to disconnect
* Use **Rclone Config** to open the Rclone configuration CLI

---

## Screenshot

![Screenshot](https://raw.githubusercontent.com/keshon/rclone-mount-manager/master/assets/screenshot.png)


## License

MIT License
