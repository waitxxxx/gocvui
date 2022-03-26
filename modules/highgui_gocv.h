#ifndef __GOCVUI_OPENCV3_HIGHGUI_H_
#define __GOCVUI_OPENCV3_HIGHGUI_H_

#ifdef __cplusplus
#include <opencv2/opencv.hpp>
extern "C" {
#endif

typedef void (*MouseCallback )(int event, int x, int y, int flags, void* param);
// Window
void Set_Mouse_Callback(const char* winname, void* on_mouse,  void* param);

#ifdef __cplusplus
}
#endif

#endif //__GOCVUI_OPENCV3_HIGHGUI_H_
