# Prototype chroot based custom composition functions

As part of `Custom Compositions` feature we're prototyping a few different approaches to how we might run "XRM functions" 
(a slightly more opinionated superset of KRM functions).

In this POC repository, before, we implemented `init-container` and `CronJob` runner. As another alternative, chroot
based runner was implemented.

The main steps of this runner:

- Read ResourceList input from Stdin. Example input was located in `ROOT/resources/resourceList.yaml`
- Get array xrm images as input of runner. Two custom images were implemented for this prototype: 
  - `sergenyalcin10/set-label-xrm:1.0-alpha` 
  - `sergenyalcin10/set-annotation-xrm:2.0-alpha`
- Pull image
- Create a directory `/chroot/image-hash`
- Extract image to the created directory
- Prepare working environment to run chroot command.
  - For running chroot command, we need /bin and /lib folders. In the custom images there are not these directories. So,
  after extraction, in working directory we have not a 	convenient environment.
  Providing this environment these to directories will be copied to the working directory:
    - `cp -r /bin /chroot/image-hash`
    - `cp -r /lib /chroot/image-hash`
- Determine image entrypoint from the Image object.
  - As you know, we have a problem about determining the executable path of the image. By using this running method,
  we have an Image object. So we can easily reach the metadata of the image. Then the entrypoint of image (path of the
  function) can be determined:
    - ```go
      cf, err := image.ConfigFile()
            if err != nil {
            panic(err)
      }
      return cf.Config.Entrypoint
      ```
- Create the `Cmd` object
- Set `cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: "/chroot/image-hash"}`
- Set `cmd.Stdin`, `cmd.Stdout` and `cmd.Stderr`
- Call `cmd.Run()`
- Before doing these steps for another XRM image, set `cmd.Stdout` of this iteration to the `cmd.Stdin` of next iteration
By doing this, the sequential running is supported.

After running of the runner, the mutated ResourceList was dumped to Stdout and also printed. For now, the output can be
debugged by checking the pod logs. However, in real scenarios, processing and using the mutated ResourceList from Stdout 
is possible.

## Custom Images

Two custom images were implemented to use for this POC.

- `sergenyalcin10/set-label-xrm:1.0-alpha`: Adds the following key-value pair as metadata.labels:
  - `"custom-composition-label"`: `"poc-label"`
- `sergenyalcin10/set-annotation-xrm:2.0-alpha`: Adds the following key-value pair as metadata.annotations:
  - `"custom-composition"`: `"poc"`

## Testing

There are three chroot based pod manifests:
- `ROOT/manifests/chroot-pod-label.yaml`: Apply only set-label image
- `ROOT/manifests/chroot-pod-annotation.yaml`: Apply only set-annotation image
- `ROOT/manifests/chroot-pod-label-annotation.yaml`: Apply both custom images

By deploying these manifests to the Kubernetes cluster, the end-to-end pipeline can easily be tested.

## Results

- The problem that we have thought about before, function executables do not have an exact path and this path needs to be
determined, has been solved because we have image metadata in the solution method.
- Potential issues with image pulling have not been addressed in this prototype. For example, some credentials may need to
be injected when trying to pull an image from a private image repo. Therefore, it will be necessary to evaluate such
situations for real scnearios.
- Since the relevant XRM functions will be run repeatedly, there will be things like pulling and extracting images each time.
In this case, it may be necessary to implement a cache management.