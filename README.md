# custom-composition-poc

This repository contains the POC implementation for [Custom Composition]
feature in [Crossplane] project.

[kpt] tool is used for rendering the resources with custom krm functions.

### Functions

In this POC, there are two custom krm functions that are implemented in Go.
First one sets annotations and the second one sets labels for resource manfiests.
The key and values of the annotations/labels are determined as a kpt command option.

The implementations of krm functions are located in `annotations` and `labels` folders.
For each function we have a `main.go`, and this main files are built in `Dockerfile`.

### Manifests

Example resource manifests are located in `resources` directory. The manifests will be
affected after the krm functions applied.

### Init Containers

All pipeline works in only one pod. This pod has three init containers:

- prep-resources: Downloads the resources and stores these resources
in ephemeral storage of the pod. This ephemeral storage was used for the input output 
pipeline.
- set-first-annotation: Sets `testAnnotationValue` annotation value for the `testAnnotationKey`
key.
- set-first-label: Sets `testLabelValue` label value for the `testLabelKey` key.

### Containers

After pipeline is done, the debugging container works. This has a sleep command for
debugging and validating the operations are successfully completed.

Container:
- debug-resources: This is used for debugging.

### Results

- As known, init containers work in a specified order and the next one waits for the
completion of previous. This is a requirement for our feature. User can control the order of
functions and manipulates output for business requirements.
- For input/output pipeline, the ephemeral storage of the pods are used. For this scenario,
this approach is successful. (ConfigMaps were not used for POC.)
- It seems that `kustomize` libraries are very talented for writing the custom krm functions.
So for some custom functions and supports the library will be very useful.
- As a next step a real-life scenario can be added for this POC repository e.g.
conditional manipulation.

[Custom Composition]: https://github.com/crossplane/crossplane/issues/2524
[Crossplane]: https://github.com/crossplane/crossplane
[kpt]: https://kpt.dev/book/04-using-functions/