FROM busybox

MAINTAINER mozhuli <weidonglee27@gmail.com>

ADD _output/kubestack /opt/cni/bin/kubestack
ADD deployment/kubestack/install-cni.sh /install-cni.sh

ENV PATH=$PATH:/opt/cni/bin
VOLUME /opt/cni
WORKDIR /opt/cni/bin
CMD ["/opt/cni/bin/kubestack"]
