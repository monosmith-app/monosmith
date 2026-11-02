// Package functions scaffolds a Python serverless function app for Azure
// Functions (the v2 programming model), AWS Lambda or Google Cloud (Cloud Run)
// functions. The host, the trigger and the runtime are choices (ADR 0072); this
// package knows them, writes the files of an app, finds a host again from its
// files, and gives such an app its start, build and package tasks.
package functions
