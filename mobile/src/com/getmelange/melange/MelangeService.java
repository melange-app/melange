package com.getmelange.melange;

import go.Go;
import go.melange.Melange;

import android.app.Service;
import android.content.Intent;
import android.content.Context;

import com.getmelange.melange.R;

import android.os.IBinder;

import android.util.Log;

import android.app.Notification;
import android.app.NotificationManager;

import java.lang.Thread;

public class MelangeService extends Service {
    boolean isRunning = false;
    NotificationManager mNotificationManager;

    @Override
    public void onCreate() {
        super.onCreate();

        // Load the Notification Manager
        mNotificationManager = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
    }
    
    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if(isRunning) {
            Log.d("MelangeService", "Not restarting Go server.");
            return Service.START_NOT_STICKY;
        }

        isRunning = true;

        // Launch Golang Application Server
        Go.init(getApplicationContext());
        try {
            new Thread(new Runnable() {
                    public void run() {
                        loadNotificationManager();
                        Log.d("MelangeService", "Finished with Notification Manager.");
                    }
                }).start();

            Melange.Run(7776, getFilesDir().getAbsolutePath(), "0.1", "android");
        } catch (Exception e) {
            Log.e("MelangeService", "Something terrible happened." + e.getMessage());
        }

        Log.d("MelangeService", "Melange Running...");

        return Service.START_NOT_STICKY;
    }

    public void createNotification(String title, String body, String id) {
        Notification theNote = new Notification.Builder(getApplicationContext())
            .setContentTitle(title)
            .setContentText(body)
            .setSmallIcon(R.drawable.small_icon)
            .build();
        mNotificationManager.notify(id, 0, theNote);
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public void onDestroy() {
        isRunning = false;
        Log.d("GoStdio", "Destroying Service");
    }

    private native void loadNotificationManager();
}
