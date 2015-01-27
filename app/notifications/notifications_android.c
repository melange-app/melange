// +build android

#include <stdlib.h>
#include "_cgo_export.h"
#include <android/log.h>
#include "jni.h"

static JNIEnv *melange_env;
static jobject melange_service;

JNIEXPORT void JNICALL
Java_com_getmelange_melange_MelangeService_loadNotificationManager (JNIEnv* env, jobject s) {
    melange_service = s; //(*env)->NewGlobalRef(env, s);
    melange_env = env;

    __android_log_print(ANDROID_LOG_DEBUG, "MelangeNative", "Starting Go Notification Thread");
    createNotificationThread();
}

// title, body, id
void createNotification(char* title, char* body, char* id) {
    __android_log_print(ANDROID_LOG_DEBUG, "MelangeNative", "Creating Notification.", melange_env, melange_service);
    JNIEnv *env = melange_env;

    // Get the method signature to fire notifications
    jclass cls = (*env)->GetObjectClass(env, melange_service);
    jmethodID mid = (*env)->GetMethodID(
        env, cls, "createNotification", 
        "(Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)V");
    if (mid == 0)
        return;

    // Get Java Strings
    jvalue jArgs[3];
    // Title
    jArgs[0] = (*env)->NewStringUTF(env, title);
    // Body
    jArgs[1] = (*env)->NewStringUTF(env, body);
    // ID
    jArgs[2] = (*env)->NewStringUTF(env, id);

    (*env)->CallVoidMethodA(env, melange_service, mid, jArgs);
}
