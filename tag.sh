#!/bin/bash

function check_status() {
  if [ $? -ne 0 ]; then
      echo "[ERROR]: Command failed!"; exit 1;
  fi
}

function set_tag() {
  local TAG=$1

  git tag -a "$TAG" -m "Repository version $TAG" || check_status
  echo "[INFO]: Tag $TAG was successfully set!"

  git push origin "$TAG" || check_status
  echo "[INFO]: Tag $TAG was successfully pushed to origin master!"
}

function increment_tag() {
  local TAG=$1
  local MAX_VNUM=999

  #Replace . with space and split into an array
  local VERSION_ARR=(${TAG//./ })

  #Get number parts and increase last one by 1
  #[FORMAT]: v0.0.0 - vN.999.999 
  local VNUM1=${VERSION_ARR[0]}
  local VNUM2=${VERSION_ARR[1]}
  local VNUM3=${VERSION_ARR[2]}
  
  # Removing v-symbol
  VNUM1="${VNUM1:1}"

  VNUM3=$((VNUM3+1))
  if [ "$VNUM3" -gt "$MAX_VNUM" ]; then
    VNUM3=0
    VNUM2=$((VNUM2+1))
    if [ "$VNUM2" -gt "$MAX_VNUM" ]; then
      VNUM2=0
      VNUM1=$((VNUM1+1))
    fi
  fi

  #Creating and assigning new tag
  NEW_TAG="v$VNUM1.$VNUM2.$VNUM3"
}

# IMPORTANT: The script supports vX.X.X tag pattern only!

INIT_TAG="v0.0.0"
NEW_TAG=""
GIT_LATEST_TAG=$(git ls-remote --tags --refs -q | sed -E 's/^[[:xdigit:]]+[[:space:]]+refs\/tags\/(.+)/\1/g' | sort -V | tail -1)

if [ -z "$GIT_LATEST_TAG" ]; then
  echo "[INFO]: No git tags found! Setting up initial tag $INIT_TAG ..."
  NEW_TAG=$INIT_TAG
  set_tag $NEW_TAG
else
  echo "[INFO]: Existing tags found! The latest is: $GIT_LATEST_TAG"
  increment_tag "$GIT_LATEST_TAG"
  echo "[INFO]: Generated new tag: $NEW_TAG"
  set_tag $NEW_TAG
fi
