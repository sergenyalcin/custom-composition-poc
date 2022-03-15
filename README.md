# custom-composition-poc

This repository contains the POC implementation for [Custom Composition]
feature in [Crossplane] project.

[kpt] tool is used for rendering the resources with custom krm functions.

## PART 1

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

All pipeline works in only one pod. This pod has four init containers:

- prep-resources: Downloads the resources and stores these resources
in ephemeral storage of the pod. This ephemeral storage was used for the input output 
pipeline.
- set-annotation: Sets `value` annotation value for the `key`
key.
- set-first-label: Sets `poc` label value for the `app` key.
- set-second-label: Sets `demo` label value for the `release` key.

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

## PART 2

In the second part of this POC, the same custom KRM functions are used. In the first part,
both two functions were located in the same image. Now, two custom KRM function images were
built:

- sergenyalcin10/set-annotations:1.0-alpha
- sergenyalcin10/set-labels:1.0-alpha

In second part, a standalone controller was implemented that takes a CRD with an input K8s
resource manifest and array of the KRM functions. Then the controller reconciles the CRD by 
running the KRM functions by using the kpt tool. For running the KRM functions, kubernetes pods
were used.

To put it simply, the operations done through the pod manifest in the first part were done 
using a controller and crd. The basic concept is the same.

The Part 2 was located in poc-controller directory. CRD manifest and examples are in the config
sub-directory.

### Results - Future Work

- Currently, the input of KRM functions is embedded in the spec of the corresponding CR. This 
input is transferred to a file in the first init container (prep-resource) and processed on it. 
Other methods can be developed to get the input and the user can use what she wants. Even the use 
of existing resources in the cluster as inputs can even be supported.

- Like part one, output locates under the /data directory in the container. And the pod is left 
running to validate the operation. However, a method has not been developed to export this output 
from the relevant container. The next step may be related to this method. Here, the issue of 
stdin/stdout, which we have been talking about since the beginning, has come to the fore again.

- The currently working controller is pretty straightforward. It is assumed that the values in the 
spec are immutable. That is, if the relevant CR is edited and the image is changed, no action is 
taken. This may be one of the points to be discussed.

- When creating containers in which KRM functions will run, it is assumed that the file to be executed 
is in a certain path (/function). There is already an issue on this subject. There is also a comment from Nic. 
Please see: https://github.com/GoogleContainerTools/kpt/issues/2567#issuecomment-1056010936

### Using CronJobs

As an extension, a new controller was added to the POC. It is located in the following path:
`poc-controller/controllers/customcomposition_cj_controller.go`

When you change the `main.go:81`:

from -> `if err = &controllers.CustomCompositionReconciler`

to -> `if err = &controllers.CustomCompositionCjReconciler`

you can use the new controller.

This controller provisions a CronJob instead of a Pod. 
Related issue: https://github.com/crossplane/crossplane/issues/2959

I didn't observe a very clear advantage or disadvantage of generating a CronJob instead of pod. 
However, as I mentioned before (in public doc), it seems more beneficial to us in terms of seeing the whole flow. 
For example, I think it would be much easier to observe operations through CronJob and make the 
relevant checks, instead of tracking individual pods created in a certain period of time.

Actually, in action, functions work in init containers fot both ways. However, using an abstraction like 
CronJob at the higher level seems to be very useful, especially in operational terms, for the reasons I 
mentioned above.

As you know, it is Kubernetes' suggestion to carry out the work to be done in a certain period and schedule 
through CronJobs and the CronJob resource has been created for this exact purpose and reason.

As a result, I think that the use of CronJob would be more appropriate for the relevant situation.

[Custom Composition]: https://github.com/crossplane/crossplane/issues/2524
[Crossplane]: https://github.com/crossplane/crossplane
[kpt]: https://kpt.dev/book/04-using-functions/