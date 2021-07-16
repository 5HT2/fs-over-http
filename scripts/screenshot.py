# Usage (requires python3):
# python screenshot.py                ## take a selection screenshot
# python screenshot.py -a             ## take a screenshot of the active window
# python screenshot.py -m             ## take a screenshot of the active monitor
# python screenshot.py "" --fancyurl  ## An optional second arg to use a shorter fancy URL type

from dotenv import load_dotenv
load_dotenv()
from datetime import datetime
import os
import pyperclip
import random
import requests
import string
import subprocess
import sys
import time

CDN_NAME = "frogg.ie"
BASE_URL = "https://{}/".format(CDN_NAME)
FOLDER_P = "https://i.l1v.in/public/i/"
S_FORMAT = "-region"
FILENAME = datetime.now().strftime("%Y-%m-%d-%H:%M:%S.png")
FILEPATH = os.environ.get('HOME') + "/pictures/screenshots/" + FILENAME

# Run bash command
def handle_bash_cmd(command):
  process = subprocess.Popen(command, stdout=subprocess.PIPE)
  output, error = process.communicate()
  
  if error is not None:
    handle_notification("Error Saving", error, "state-error")
    exit(1)

# Send a notification
def handle_notification(title, description, icon):
  bashCmd = "notify-send|" + title + "|" + description + "|--icon=" + icon + "|--app-name=" + CDN_NAME
  handle_bash_cmd(bashCmd.split("|"))

# Generates a random 3 long + .png filename
def get_file_name():
  # Choose from "A-z0-9"
  letters = string.ascii_letters + string.digits
  result_str = ''.join(random.choice(letters) for i in range(3))
  return result_str + ".png"

# Returns the status code for the current https://frogg.ie/`file_name` url
def get_status_code():
  return requests.get(BASE_URL + file_name).status_code

# If the second arg of the script is set, change S_FORMAT
if len(sys.argv) > 1:
  S_FORMAT = sys.argv[1]

# Take screenshot with spectacle
bashCmd = "spectacle " + S_FORMAT + " -p -b -n -o=" + FILEPATH + " >/dev/null 2>&1"
handle_bash_cmd(bashCmd.split())

# Wait for spectacle to save the file
while os.path.isfile(FILEPATH) is False:
  time.sleep(0.2)

# Get initial filename and status
file_name = get_file_name()
status_code = get_status_code()

# Loop until the filename isn't taken
while status_code == 200:
  file_name = get_file_name()
  status_code = get_status_code()

# Be sure it's a 404 Not Found
if status_code == 404:
  # Upload file
  files = {'file': open(FILEPATH, 'rb')}
  response = requests.post(FOLDER_P + file_name, files=files, headers={'Auth': os.environ.get("FOH_SERVER_AUTH")})

  if response.status_code != 200:
    handle_notification("Error " + str(response.status_code), "Response: " + response.headers["X-Server-Message"], "state-error")
    exit(1)

  pyperclip.copy(BASE_URL + file_name)
  handle_notification("Saved screenshot", file_name, "spectacle")
else:
  handle_notification("Error " + str(response.status_code), "Response: " + response.headers["X-Server-Message"], "state-error")
