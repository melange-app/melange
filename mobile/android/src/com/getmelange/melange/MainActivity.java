/*
 * Copyright 2014 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package com.getmelange.melange;

import com.getmelange.melange.MelangeService;

import android.app.Activity;
import android.os.Bundle;
import android.content.Context;

import android.content.Intent;

import android.os.Build;
import android.content.pm.ApplicationInfo;

import android.util.Log;

import android.view.Menu;

import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.webkit.WebSettings;


/*
 * MainActivity is the entry point for the melange app.
 *
 * From here, the Go runtime is initialized and a Go function is
 * invoked via gobind language bindings.
 *
 * See example/libhello/README for details.
 */
public class MainActivity extends Activity {
    WebView webContent;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // Enable Debugging for the WebView
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
            if (0 != (getApplicationInfo().flags &= ApplicationInfo.FLAG_DEBUGGABLE))
            { WebView.setWebContentsDebuggingEnabled(true); }
        }

        webContent = new WebView(this);

        WebSettings webSettings = webContent.getSettings();
        webSettings.setJavaScriptEnabled(true);

        webContent.setWebViewClient(new WebViewClient() {
                @Override
                public void onPageFinished(WebView view, String url) {
                    Log.d("Melange/GoStdio", "Page is finished loading!");
                }
            });

        setContentView(webContent);

        if (savedInstanceState == null) {
            Intent intent = new Intent(this, MelangeService.class);
            startService(intent);

            webContent.loadUrl("http://app.melange.127.0.0.1.xip.io:7776/Index.html#startup");
        } else {
            webContent.restoreState(savedInstanceState);
        }
    }

    @Override
    protected void onSaveInstanceState(Bundle outState) {
        webContent.saveState(outState);
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        return false;
    }

    @Override
    public void onLowMemory() {
        Log.d("Melange", "NO MEMORY LEFT! BAIL OUT!");
    }

    @Override
    public void onTrimMemory(int level) {
        Log.d("Melange", "STARTING A MEMORY TRIM!");
    }
}
