# TODO

## GitHub

- [ENHANCEMENT] Use changesets to manage changes
- [ENHANCEMENT] GitHub action to build and release when a tag is pushed
  
## General

- [ENHANCEMENT] Update goreleaser to build additional versions
- [ENHANCEMENT] Rewrite the README
  - Add section on how to override the PORT and HOST (`VITE_HOST=xxxx VITE_PORT=xxxx pnpm run dev`)
- [ENHANCEMENT] Icon for MAC
- [ENHANCEMENT] Add dockerfile and docker compose

## UI

### General

- [ENHANCEMENT] Add completed tick to courses
- [ENHANCEMENT] Theme (https://ui.jln.dev)
- [ENHANCEMENT] Update query param to all pages like settings -> courses/tags/logs etc as the uses filters
- [ENHANCEMENT] Add search (https://discord.com/channels/1116682155809067049/1117779396992979024/1163925360228962385)
- [ENHANCEMENT] Change how frequently the course availability check is run
- [ENHANCEMENT] Support for FFMPEG path
- [ENHANCEMENT] On mobile use a drawer for tags
- [ENHANCEMENT] Write a general course scanner 
  - Add 1 or more scans, do a bulk query for all in the list
  - take a writable and update the status

### Home

- [ENHANCEMENT] Hide ongoing when there are no ongoing courses
- [ENHANCEMENT] Support adding categories from on the home page
- [ENHANCEMENT] Fix the difference in location of the loading icon and the error
- [ENHANCEMENT] Change from carousel to no carousel
- [ENHANCEMENT] Add completed and updated icons on course cards

### Courses

- [ENHANCEMENT] Rework filters to use shadcn drawer on mobile
- [ENHANCEMENT] Additional filter for favorites

### Course

- [ENHANCEMENT] When a course is unavailable, show a message saying unavailable
- [ENHANCEMENT] Support PDF
- [ENHANCEMENT] Show 'scanning' loading page when a course is first added and scanning is in progress'
- [ENHANCEMENT] Rework menu for large and small to use the same content (instead of duplicating)

#### Video

- [BUG] When video becomes unavailable, toast renders again and again

#### Mobile

- [BUG] Sometimes the menu opens and is empty. A quick scroll fixes things
  - Fixed?

### Settings

#### General

- [ENHANCEMENT] Mark a course as complete / reset progress

#### Courses

- [ENHANCEMENT] The scan status should show for at least 1 second (scan then rescan to test)
- [ENHANCEMENT] Filters
- [ENHANCEMENT] Support the action to `move` multiple courses to another path
- [ENHANCEMENT] Add action to set courses as favourite

#####  Add

-

##### Details

- [ENHANCEMENT] Add move button and file system popup for relocating a course (and assets)
- [ENHANCEMENT] Allow changing the course card from the UI
- [ENHANCEMENT] Mark a course as complete / reset progress
- [ENHANCEMENT] Rename a file
- [ENHANCEMENT] Add button to set/remove as favorite
- [ENHANCEMENT] Support course synopsis
- [ENHANCEMENT] Support alt name for course


#### Tags

- [ENHANCEMENT] Add courses to tag(s)

#### Logs

- [ENHANCEMENT] Filter by data
- [ENHANCEMENT] Auto refresh?

## Backend

### General

- [ENHANCEMENT] Remove -ST1003 from audit
- [ENHANCEMENT] Use mattn sqlite3 driver
- [ENHANCEMENT] Allow settings course as favourite

### API

- [ENHANCEMENT] Support moving a course to another path (along with assets)
- [ENHANCEMENT] Mark a course as complete / reset progress
- [ENHANCEMENT] Rename a file

### Cron

- [ENHANCEMENT] Removing logs after n days

### Tags

- [ENHANCEMENT] Analyze and optimize the DB

### Course Scan

- [BUG] After scanning a course, run a course refresh incase new assets were added, removed 

### Course Scanner

- [ENHANCEMENT] Batch adding assets and attachments
