# TODO

## GitHub

- [ENHANCEMENT] Use changesets to manage changes
- [ENHANCEMENT] GitHub action to build and release when a tag is pushed
  
## General

- [ENHANCEMENT] Update goreleaser to build additional versions
- [ENHANCEMENT] Rewrite the README
- [ENHANCEMENT] Icon (favicon, application icon)

## UI

### General

- [BUG] Building without .env file (pnpm run build)
- [BUG] When loading images, the spinner can stop but there is a delay before the image is shown
- [ENHANCEMENT] theme (https://ui.jln.dev)
- [ENHANCEMENT] Update query param to all pages like settings -> courses/tags/logs etc as the uses filters
- [ENHANCEMENT] Add search (https://discord.com/channels/1116682155809067049/1117779396992979024/1163925360228962385)
- [ENHANCEMENT] Change how frequently the course availability check is run
- [ENHANCEMENT] Support for FFMPEG path
- [ENHANCEMENT] On mobile use a drawer for tags

### Home

- [BUG] New courses timestamp should be created_at
- [BUG] Update px to match header when screen size is xs or sm
- [ENHANCEMENT] Hide ongoing when there are no ongoing courses
- [ENHANCEMENT] Get image for landing page
- [ENHANCEMENT] Support adding categories from on the home page

### Courses

- [ENHANCEMENT] Rework filters to use shadcn drawer on mobile

### Course

- [ENHANCEMENT] When a course is unavailable, show a message saying unavailable
- [ENHANCEMENT] Support PDF
- [ENHANCEMENT] Show 'scanning' loading page when a course is first added and scanning is in progress'
- [BUG] Hover over details icon in menu. The tooltip flickers open and closed and part of the menu appears behind

#### Video

- [BUG] When video becomes unavailable, toast renders again and again

#### Mobile

- [BUG] Sometimes the menu opens and is empty. A quick scroll fixes things

### Settings

#### General

-

#### Courses

- [ENHANCEMENT] The scan status should show for at least 1 second (scan then rescan to test)
- [ENHANCEMENT] Filters
- [ENHANCEMENT] Support the action to `move` multiple courses to another path

#####  Add

- [BUG] The mobile drawer slider (line at the top) does not show when scrolling is in play
- [BUG] On mobile, show the toast at the top of the screen

##### Details

-  [ENHANCEMENT] Add move button and file system popup for relocating a course (and assets)
-  [ENHANCEMENT] Allow changing the course card from the UI

#### Tags

- [BUG] Adding the same tag with different case (upper/lower/capital)
- [ENHANCEMENT] Add courses to tag(s)

#### Logs

- [ENHANCEMENT] Filter by data
- [ENHANCEMENT] Auto refresh?

## Backend

### General

- [ENHANCEMENT] Remove -ST1003 from audit
- [ENHANCEMENT] Use mattn sqlite3 driver

### API

- [ENHANCEMENT] Support moving a course to another path (along with assets) 

### Cron

- [ENHANCEMENT] Removing logs after n days

### Tags

- [ENHANCEMENT] Currently uppercase and lowercase tags are different and so uppercase are ordered first. Make them case insensitive
- [ENHANCEMENT] Analyze and optimize the DB

### Assets and Attachments

- 

### Course Scanner

- [ENHANCEMENT] Batch adding assets and attachments
