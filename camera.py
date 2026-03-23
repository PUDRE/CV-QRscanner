import cv2
import numpy as np
import sys
import time

class QRScannerOpenCV:
    def __init__(self):
        self.cap = None
        self.qr_detector = cv2.QRCodeDetector()
        self.scanned_ids = set()
        self.last_scanned_time = 0
        self.scan_cooldown = 2.0
        
    def initialize_camera(self):
        for camera_index in range(0, 3):
            self.cap = cv2.VideoCapture(camera_index, cv2.CAP_DSHOW)
            if self.cap.isOpened():
                ret, frame = self.cap.read()
                if ret:
                    print(f"Camera {camera_index} opened successfully")
                    self.cap.set(cv2.CAP_PROP_FRAME_WIDTH, 1280)
                    self.cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 720)
                    self.cap.set(cv2.CAP_PROP_FPS, 30)
                    return True
                else:
                    self.cap.release()
        
        print("ERROR: Failed to open any camera")
        return False
    
    def decode_qr_opencv(self, frame):
        try:
            retval, decoded_info, points, straight_qrcode = self.qr_detector.detectAndDecodeMulti(frame)
            if retval and len(decoded_info) > 0:
                for i, data in enumerate(decoded_info):
                    if data and len(data) > 0:
                        try:
                            student_id = int(data.strip())
                            if student_id > 0:
                                points_array = points[i].astype(int)
                                return student_id, points_array
                        except ValueError:
                            continue
        except Exception as e:
            pass
        return None, None
    
    def process_frame(self, frame):
        """Process frame and look for QR codes"""
        display_frame = frame.copy()
        found_id = None
        qr_points = None
        height, width = display_frame.shape[:2]
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8,8))
        enhanced = clahe.apply(gray)
        enhanced_color = cv2.cvtColor(enhanced, cv2.COLOR_GRAY2BGR)
        student_id, points = self.decode_qr_opencv(enhanced_color)
        if student_id is not None and points is not None:
            found_id = student_id
            qr_points = points
            n = len(points)
            for j in range(n):
                cv2.line(display_frame, tuple(points[j]), tuple(points[(j+1) % n]), (0, 255, 0), 3)
            center_x = int(np.mean(points[:, 0]))
            center_y = int(np.mean(points[:, 1]))
            cv2.circle(display_frame, (center_x, center_y), 8, (0, 0, 255), -1)
            text = f"ID: {student_id}"
            text_size = cv2.getTextSize(text, cv2.FONT_HERSHEY_SIMPLEX, 0.7, 2)[0]
            text_x = center_x - text_size[0] // 2
            text_y = center_y - 20
            cv2.rectangle(display_frame, 
                        (text_x - 5, text_y - text_size[1] - 5),
                        (text_x + text_size[0] + 5, text_y + 5),
                        (0, 0, 0), -1)
            cv2.putText(display_frame, text, 
                      (text_x, text_y),
                      cv2.FONT_HERSHEY_SIMPLEX, 0.7, (0, 255, 255), 2)
        return display_frame, found_id
    
    def scan_real_time(self):
        """Main real-time scanning loop"""
        if not self.initialize_camera():
            return
        print("Scanning started.")
        print("Press ESC to finish scanning")
        print("Press SPACE for manual student input")
        window_name = 'QR Code Scanner - ESC to exit, SPACE for manual input'
        cv2.namedWindow(window_name, cv2.WINDOW_NORMAL)
        cv2.resizeWindow(window_name, 800, 600)
        last_scanned_id = None
        while True:
            ret, frame = self.cap.read()
            if not ret:
                print("ERROR: Failed to read frame from camera")
                break
            frame = cv2.flip(frame, 1)
            display_frame, found_id = self.process_frame(frame)
            height, width = display_frame.shape[:2]
            current_time = time.time()
            if found_id is not None:
                if last_scanned_id != found_id or (current_time - self.last_scanned_time) > self.scan_cooldown:
                    print(f"SCAN:{found_id}")
                    sys.stdout.flush()
                    last_scanned_id = found_id
                    self.last_scanned_time = current_time
                    cv2.rectangle(display_frame, (0, 0), (width, height), (0, 255, 0), 10)
            info_text = "ESC: exit | SPACE: manual input"
            cv2.putText(display_frame, info_text, (10, 30),
                       cv2.FONT_HERSHEY_SIMPLEX, 0.7, (255, 255, 255), 2)
            cv2.imshow(window_name, display_frame)
            key = cv2.waitKey(1) & 0xFF
            if key == 27:  # ESC
                print("Scanning finished by user")
                break
            elif key == 32:  # SPACE
                print("MANUAL_INPUT")
                sys.stdout.flush()
            cv2.waitKey(1)
        self.cap.release()
        cv2.destroyAllWindows()
        print("SCAN_COMPLETE")

if __name__ == "__main__":
    try:
        scanner = QRScannerOpenCV()
        scanner.scan_real_time()
    except Exception as e:
        print(f"ERROR: {str(e)}")
        import traceback
        traceback.print_exc()