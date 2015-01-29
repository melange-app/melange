//
//  MelangeBrowserProtocol.m
//  Melange
//
//  Created by Hunter Leath on 10/16/14.
//  Copyright (c) 2014 Hunter Leath. All rights reserved.
//

#import "MelangeBrowserProtocol.h"

@interface MelangeBrowserProtocol ()
@property (nonatomic, strong) NSURLConnection *connection;

@end

@implementation MelangeBrowserProtocol

+ (NSURLRequest *)canonicalRequestForRequest:(NSURLRequest *)r
{
    NSLog(@"Getting Canonical for: %@", [r URL]);
    NSMutableURLRequest *mutURL = [r mutableCopy];
    
    // Check for Melange
    NSString *host =[[r URL] host];
    if([host hasSuffix:@".melange"]) {
        NSString *newURLString = [NSString stringWithFormat:@"%@://%@%@", [[r URL] scheme], @"localhost:7776", [[r URL] path]];
        NSURL *newURL = [NSURL URLWithString:newURLString];
        [mutURL setURL:newURL];
        [mutURL setValue:host forHTTPHeaderField:@"Host"];
        NSLog(@"Changed %@ to %@", host, mutURL);
    }
    
    [NSURLProtocol setProperty:@YES forKey:@"MelangeHandled" inRequest:mutURL];
    
    return mutURL;
}

+ (BOOL)canInitWithRequest:(NSURLRequest *)r {
    if ([NSURLProtocol propertyForKey:@"MelangeHandled" inRequest:r] != nil)
        return NO;
    
    return YES;
}

- (void)startLoading
{
    NSLog(@"Start loading for %@", [[self.request URL] host]);
    self.connection = [NSURLConnection connectionWithRequest:self.request delegate:self];
}

- (void)stopLoading
{
    [self.connection cancel];
}

- (void)connection:(NSURLConnection *)connection didReceiveData:(NSData *)data
{
    [self.client URLProtocol:self didLoadData:data];
}

- (void)connection:(NSURLConnection *)connection didFailWithError:(NSError *)error
{
    [self.client URLProtocol:self didFailWithError:error];
    self.connection = nil;
}

- (void)connection:(NSURLConnection *)connection didReceiveResponse:(NSURLResponse *)response
{
    [self.client URLProtocol:self didReceiveResponse:response cacheStoragePolicy:NSURLCacheStorageAllowed];
}

- (void)connectionDidFinishLoading:(NSURLConnection *)connection
{
    [self.client URLProtocolDidFinishLoading:self];
    self.connection = nil;
}

@end
