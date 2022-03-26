#include ".highgui_gocv.h"

void Set_Mouse_Callback(const char* winname, void* on_mouse,  void* param) {
    cv::setMouseCallback(winname, (CvMouseCallback)on_mouse,param);
}