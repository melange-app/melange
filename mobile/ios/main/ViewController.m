//
//  ViewController.m
//  Melange
//
//  Created by Hunter Leath on 10/16/14.
//  Copyright (c) 2014 Hunter Leath. All rights reserved.
//

#import "ViewController.h"

@interface ViewController ()

@end

@implementation ViewController {
    UIView *loadingView;
}

- (void)viewDidLoad {
    [super viewDidLoad];
    // Do any additional setup after loading the view, typically from a nib.
    
    NSLog(@"%@", [[NSURL URLWithString:
           @"http://app.melange:7776/Index.html"] host]);
    [webView  loadRequest:
                [NSURLRequest requestWithURL:
                                [NSURL URLWithString:
                                    @"http://app.melange:7776/Index.html#startup"]]];
    webView.delegate = self;
    [self createActivityIndicator];
}

- (void)webViewDidStartLoad:(UIWebView *)webView {
    [loadingView setHidden:NO];
}

- (void)webViewDidFinishLoad:(UIWebView *)webView {
    [loadingView setHidden:YES];
}

- (void) createActivityIndicator {
    loadingView = [[UIView alloc]initWithFrame:CGRectMake(100, 400, 80, 80)];
    loadingView.backgroundColor = [UIColor colorWithWhite:0. alpha:0.6];
    loadingView.layer.cornerRadius = 5;
    
    UIActivityIndicatorView *activityView=[[UIActivityIndicatorView alloc] initWithActivityIndicatorStyle:UIActivityIndicatorViewStyleWhiteLarge];
    activityView.center = CGPointMake(loadingView.frame.size.width / 2.0, 35);
    [activityView startAnimating];
    activityView.tag = 100;
    [loadingView addSubview:activityView];
    
    UILabel* lblLoading = [[UILabel alloc]initWithFrame:CGRectMake(0, 48, 80, 30)];
    lblLoading.text = @"Loading...";
    lblLoading.textColor = [UIColor whiteColor];
    lblLoading.font = [UIFont fontWithName:lblLoading.font.fontName size:15];
    lblLoading.textAlignment = NSTextAlignmentCenter;
    [loadingView addSubview:lblLoading];
    
    [self.view addSubview:loadingView];
}

- (void)didReceiveMemoryWarning {
    [super didReceiveMemoryWarning];
    // Dispose of any resources that can be recreated.
}

@end
