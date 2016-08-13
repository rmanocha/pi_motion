#!/bin/bash
#echo "create table last_motion (rid integer primary key, start_time datetime, end_time datetime); insert into last_motion SELECT 1, datetime('now', '-5 seconds');" | sqlite3 last_motion.db
echo "create table last_motion (rid integer primary key, start_time datetime, end_time datetime);" | sqlite3 last_motion.db

