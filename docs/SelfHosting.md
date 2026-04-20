# Kuberneties Basicks
Kuberneties is a platform that is desinged to be scalable so it is a little difrent then your run of a mill docker stack. the main difrance is the way kuberneties handles work loads. In a kuberneties cluster you should have both A Controll Plane, and (a or many) Worker Nodes. The controll plane is what handles the dispatching of the work and the worker nodes dose all the work.
# machine infraestructure
I (Landon) virtualise all of my kuberneties nodes, at the time i was deploying this app i had only two worker nodes and one controll plane do whatever you plese just keep in mind the more machines u have the better u can push out more replicas for smother sailing. Now when setting up our vm inferstructure i recomend using IAC(infrastructure as code) i have my homelabs infrastructure published down below to use as an example.
[Skoxie-iac Repo](https://codeberg.org/Woxie/skoxie-iac)

# Where do i get the Images?
U(points at u) future me.. will have to push a docker image to docker hub or your own docker registry and then use that with a helm chart. I use Artifact hub to publish my completed charts so that i can pull them into my argo cd infrastructure

# the helm repo
the helm chart repo is hosted on a github page evrytime you want to update the charts plese update the vershion number and push to main thanks