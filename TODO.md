# TODO

## General

- [ ] Update makefile to use goreleaser
- [ ] Rewrite the README

## UI

### General

- [x] Use await
- [x] Tidy Loading component
- [x] Fix selecting video assets then moving forward and backward. There is an issue with history
- [x] Mobile menu clips header
- [x] Disable scrolling when mobile menu is open
- [x] Fix time in the UI. Newly added courses show in an hour ... VPN?
- [x] When loading course card spin the fallback icon
- [ ] Fix theme
- [ ] Update query param to all pages like settings -> courses/tags/logs etc as the uses filters
- [x] On scan, have a minimum 'load' of 1 second to stop flickering
- [ ] Add search (https://discord.com/channels/1116682155809067049/1117779396992979024/1163925360228962385)
- [ ] Change how frequently the course availability check is run
- [ ] Support for FFMPEG path
- [x] Move from lucide to phosphor icons
- [ ] Test building without .env file (pnpm run build)


### Home

- [x] Fix Load More border in carousel
- [x] Landing page when there are no courses added
- [ ] Hide ongoing when there are no ongoing courses
- [x] New courses timestamp should be created_at
- [ ] Get image for landing page

#### Categories

- [ ] Click a button and select 1 or more tags to create a category
- [ ] Name the category
- [ ] Show categories on the home page

### Courses

- [x] https://dribbble.com/shots/23132040-E-learning-website-course-details
- [x] Filter
- [x] When a filter is selected, make sure you are on page 1 (if not already)
- [x] Fix timestamp shown in card. It is not showing the updated_at time
- [x] Fix card when only 1 is showing. It goes small for some reason
- [ ] Rework filters to use shadcn drawer on mobile

### Course

- [x] When a course is unavailable, still show the menu
- [x] When moving to the prev/next course, scroll to that item in the normal menu
- [ ] When a course is unavailable, show a message saying unavailable
- [ ] Support PDF
- [ ] Show 'scanning' loading page when a course is first added and scanning is in progress'
- [ ] Fix hover over details icon in menu. The tooltip flickers open and closed and part of the menu appears
      behind

#### Video

- [x] Settings; auto play, auto next
- [x] Add gradient to top of controls when on mobile
- [x] Store state in local storage
- [x] Rework mobile settings to use shadcn drawer
- [x] Show rewind and forward buttons on xs and sm
- [x] Add support for replay
- [x] Store the volume in local storage
- [x] Show volume control on xs, sm when the data-pointer is fine
- [x] Fix error on md+ settings. Error in console
- [x] On mobile, when the video is playing, clicking the video, when controls are hidden, should show
       the controls
- [x] When the `autoplay next` is enabled, and the time slider is dragged to the end, it sometimes loads
      next + 1. It does not happened when the video ends naturally
- [ ] When video becomes unavailable, fix toast so it doesn't show again and again and show a message on
      the video saying unavailable

#### Mobile

- [x] Add `x` to menu
- [x] Close when menu item is clicked
- [x] When opening the menu, scroll to the selected menu item
- [x] Make the prev/next buttons use `flex-col` and take up 4/5 of the left/right side
- [x] Add fade to top and bottom of menu
- [ ] Fix prev/next button. It stays highlighted after being clicked
- [ ] Sometimes the menu opens and is empty. A quick scroll fixes things

### Settings

#### General

- [x] Fix filters for mobile

#### Courses

- [ ] Filter
- [ ] Store table state in local storage/db
- [x] Add table action `Add Tags`
- [x] Use shadcn table
- [x] Fix scan updated_at time .. it should be the time of the last scan
- [ ] Fix scan then rescan. It does not show the `scan` status
- [ ] Support the action to update  multiple courses to another path. This 

#####  Add

- [x] Rework into a dialog
- [x] Rework getting all courses to be more efficient
- [x] Fix border in back button when adding courses (course selection)
- [x] Use drawer for small screens
- [ ] Fix drawer slider not showing on mobile
- [ ] fix toast when adding courses on mobile. Hides bottom of the drawer
- [x] Fix refresh (navigate into dir, navigate up, click refresh)

##### Details

- [x] Fix issue with a page refresh happening after clicking the scan button
- [x] When deleting a course, move to /settings/courses
- [x] Fix scan updated_at time .. it should be the time of the last scan
- [x] Cannot add and delete tags at the same time
- [ ] Rework size of text/icons for lg+
- [ ] Add move button and file system popup for relocating a course (and assets)

#### Tags

- [x] Add table
- [x] Allow deleting of tags
- [x] Allow adding of tags
- [x] Fix sorting by course count
- [x] Fix adding the same tag with different case (upper/lower/capital)
- [x] Allow editing of tags
- [ ] Add courses to tag(s)
- [ ] Fix adding the same tag with different case (upper/lower/capital)

#### Logs

- [x] Add table
- [x] Filters (log level, request type, etc)
- [x] Filter by data.type
- [ ] Filter by data

## Backend

### General

- [x] Every DAO should support tx
- [ ] Remove -ST1003 from audit

### API

- [ ] Support moving a course to another path (along with assets) 

### Logs

- [x] Add a separate logs DB
- [x] Update all logs (remove zerolog)
- [ ] Support removing logs after n days

### Tags

- [ ] Currently uppercase and lowercase tags are different and so uppercase are ordered first. Make them case insensitive

### Assets and Attachments

- [x] Add column for md5sum of file

### Course Scanner

- [ ] Fix finding card
