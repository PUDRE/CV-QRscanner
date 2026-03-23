# Introdaction
A pet project for automated student attendance tracking using QR code scanning via webcam.  
Written in Go (core logic) and Python (QR recognition via OpenCV).  
Real-time interaction, visual feedback, and manual ID input.
Data storage and UI are primitive.

#Technologies
Component       | Technologies / Libraries                               
----------------|-------------------------------------------------------
Go core         | `os/exec`, `bufio`, `encoding/json` (data storage)    
Python script   | `opencv-python`, `numpy`                              
Data format     | JSON (student list, attendance records)               
Communication   | Inter‑process via stdout / stdin                      

#Requirments
- Go 1.18+
- Python 3.7+
- OpenCV for Python (install via pip)
