TODO

https://dribbble.com/shots/23132040-E-learning-website-course-details

- [x] Use await
- [x] Tidy Loading component


## Page -> Settings -> Tags

- [x] Add table
- [x] Allow deleting of tags
- [x] Allow adding of tags
- [ ] Allow editing of tags
- [x] Fix sorting by course count

## Page -> Settings -> Courses -> Add

- [ ] Rework into a dialog
- [ ] Rework getting all courses to be more efficient

## Page -> Home

- [ ] Landing page when there are no courses added (use a DB flag)
- [ ] Hide ongoing when there are no ongoing courses
- [ ] Categories
  - [ ] Click a button and select 1 or more tags to create a category
  - [ ] Name the category
  - [ ] Show categories on the home page

## Page -> Course

- [ ] When a course is unavailable, still show the menu
- [ ] Support PDF
- [ ] Show progress bar with tooltip for course progress
- [ ] Fix issue when back button because a_id is updated
- [ ] Video
  -  [ ] Settings (speed change, auto play)
  -  [ ] Store state in DB
  -  [ ] Issue -> Finish video, seek to middle then play. It jumps back to start

## Page -> Settings -> General

- [ ] Change how frequently the course availability check is run
- [ ] Support for FFMPEG path

## Page -> Settings -> Courses

- [ ] Filter
- [ ] Store table state in local storage/db

## Search

- [ ] Add search (https://discord.com/channels/1116682155809067049/1117779396992979024/1163925360228962385)

## Backend -> Assets and Attachments

- [ ] Add column for md5sum of file

## Backend -> logs

- [ ] Add a separate logs DB
- [ ] Add logs