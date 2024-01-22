# Off Course (WIP)

## Description

`Off Course` is an application designed to enable viewing and managing course material locally

## Overview

- **Frontend**: SvelteKit with TypeScript
- **Backend**: Golang with SQLite3 database

## Running

To run `Off Course`, download the latest release for your platform from the [releases page]() and execute

By default, this the application will be available at `localhost:9081`

Note: You may override the port by setting the running the binary with `-port :<port>`

### Database

When first launched, `Off Course` will create a `oc_data` directory along side the binary

This directory will contain the database

On subsequent runs, this database will be used

## Usage

### Adding a Course

Courses may be added via the frontend, by navigating to `Settings` > `Courses` and clicking `Add Courses`

This will open a dialog where you may navigate your file system and select courses to add

The structure of a course on disk is important. See [Course Structure](#course-structure) for more information

After adding courses, the table on this page will automatically refresh to show the newly added courses

#### Scanning

When a course is first added, it will be scanned in order to identify assets and attachments

To see the details about a course click `...` > `Details` on a course row

Over time, you may add or remove assets/attachments via the file system. To pull in the new information about a course, you may perform a manual scan by clicking `...` > `Scan` on a course row within the table

Note: When a scan is performed, assets and attachments that have been removed from the file system will be removed from the database. As such, any progress information about these assets will be lost

#### Availability

Over time, the availability of courses may change. For example, a course may be removed from the file system, or simply renamed

`Off Course` will not delete courses from the database when they are no longer available. Instead, it will mark them as unavailable, allowing you to maintain progress information about courses that you have previously added, but are no longer available

If you wish to remove a course from the database, you may do so by clicking `...` > `Delete` on a row within the table

Note: A job runs periodically in the background to check the availability of courses. In addition to this, when a manual scan is performed, the availability of courses will be checked

## Course Structure

A course is simply a directory containing assets and attachments.

The name of the course will be the name of the directory

### Card

An image named `card.xxx` may be be placed at the root of the course directory, whereby `xxx` is a supported image extension (.jpg, .png, .webp, .tiff)

### Assets and Attachments

Assets and attachments are files within a course directory

Assets are considered primary course material, such as videos, html files and pdf files

Attachments are supplementary materials linked to assets

#### File Organization

Assets and attachments may be placed at the root of the course directory, or within subdirectories, which are seen as chapters/sections

Subdirectories within subdirectories are ignored

#### Filename

The filename breakdown for assets and attachments is extremely important

_Assets_

Assets must start with a numerical prefix, followed by a descriptive title, and finally a supported asset type extension

For example, `01 Introduction.mp4`

See [Supported Asset Types](#supported-asset-types) for a list of supported asset types

_Attachments_

Attachments are linked to assets via the numerical prefix and as such attachments must start with a numerical prefix

The prefix may optionally be followed by a descriptive title and an extension

For example, `01`, `01.zip` and `01 Introduction Notes.txt` would all become attachments of `01 Introduction.mp4`

Note: The prefix and title maybe separated by a space or a dash, for example `01 - Introduction.mp4` is also valid

_other_

Any other files will be ignored

#### Asset Priority

Assets have a priority: Video > HTML > PDF

If multiple assets are identified with the same prefix, ex. `01`, the asset with the highest priority will become the asset. All remaining assets will be downgraded to attachments

For example, if `01 Introduction.mp4` and `01 Introduction.html` both exist, `01 Introduction.mp4` will be the asset and `01 Introduction.html` will be an attachment

If multiple assets are identified with the same prefix, ex `01`, and are of the same priority, ex `video`, the first asset seen alphabetically will become the asset. All remaining assets will be downgraded to attachments

For example, if `01 Video 1.mp4` and `01 Video 2.mp4` both exist, `01 Video 1.mp4` will be the asset and `01 Video 2.mp4` will be an attachment

#### Supported Asset Types

The following extensions will be treated as assets, given the filename matches the naming structure described above

**video**
- avi
- mkv
- flac
- mp4
- m4a
- mp3
- ogv
- ogm
- ogg
- oga
- opus
- webm
- wav

**HTML**
- htm
- html

**PDF**
- pdf

### Example Structure

```
My Course/
│
├── card.jpg                    # Course card image
│
├── 01 Basics/                  # Chapter 1 content
│   ├── 01 Overview.mp4         # Main asset for Chapter 1
│   ├── 01 Overview Notes.txt   # Attachment to '01 Overview.mp4'
│   └── 02 Example.html         # Main asset (HTML file)
│
└── 02 Advanced/                # Chapter 2 content
    ├── 01 Deep Dive.pdf        # Main asset for Chapter 2
    └── 01 Source Links.txt     # Attachment to '01 Deep Dive.pdf'
...
```


## Development

### Prerequisites

**Frontend**
- Node.js >=20
- pnpm >= 8

**Backend**
- Go >= 1.20

### Running

Clone the repository

```bash
git clone https://github.com/geerew/off-course.git
cd off-course
```

During development, the frontend and backend will be started separately

**Backend**

1. Install dependencies
   ```bash
   go mod download
   ```

2. Start the backend server
   ```bash
   go run main.go
   ```

The backend will be running on `localhost:9081`

**Frontend**

1. Open a new terminal
2. Navigate into `./ui`
3. Create a `.env` file

4. Add the following to the .env file

   Note: Change the port to match the port the backend is running on, which by default is `9081`

   ```
   export PUBLIC_BACKEND=http://localhost:9081
   ```

5. Install dependencies
   ```bash
    pnpm install
   ```

6. Start the frontend
   ```bash
   pnpm run dev
   ```

The frontend will be running on `localhost:5173`

### Building

**Frontend**

1. Navigate into `./ui`
2. Build
   ```bash
   pnpm run build
   ```

The frontend will be built to `./ui/build`

**Backend**

1. Navigate into the root of the project
2. Build
   ```bash
    go build
   ```

The binary will be output as `./off-course` and will embed the frontend