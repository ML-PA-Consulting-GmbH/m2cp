FROM ubuntu:jammy as base
ENV DEBIAN_FRONTEND=noninteractive

# Run apt-get commands together to always ensure apt-get update is run
# before using apt-get install to avoid issues with stale apt caches.
RUN apt-get update -qq && \
	apt-get dist-upgrade --yes && \
	apt-get install --yes -qq --no-install-recommends \
		gnupg \
		xdelta3 \
		sudo \
		systemd \
		vim \
		ssh \
		socat \
		wget \
		iputils-ping \
		fuse3 \
		libfuse3-3 \
		htop


COPY snapd.deb .
RUN apt install --yes -qq ./snapd.deb

# Clean apt lists because we'll be copying the entire filesystem to a blank image
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists

# Ensure snapd's system-key file exists
RUN touch /var/lib/snapd/system-key

# Stop udevadm from working
RUN dpkg-divert --local --rename --add /sbin/udevadm
RUN ln -s /bin/true /sbin/udevadm

# Remove systemd 'wants' triggers
RUN rm -f \
		/etc/systemd/system/*.wants/* \
		/lib/systemd/system/local-fs.target.wants/* \
		/lib/systemd/system/multi-user.target.wants/* \
		/lib/systemd/system/sockets.target.wants/*initctl*

# Remove everything except tmpfiles setup in sysinit target
RUN find \
		/lib/systemd/system/sysinit.target.wants \
		\( -type f -or -type l \) -and -not -name '*systemd-tmpfiles-setup*' \
		-delete

# Remove UTMP updater service
RUN rm -f /lib/systemd/system/systemd-update-utmp-runlevel.service

# Disable /tmp mount
RUN rm -vf /usr/share/systemd/tmp.mount

# Disable most systemd console output
RUN echo ShowStatus=no >> /etc/systemd/system.conf

# Disable ondemand.service
#RUN systemctl disable ondemand.service

# Set basic.target as default
RUN systemctl set-default basic.target

# Enable the services we care about
RUN systemctl enable snapd.service
RUN systemctl enable snapd.socket
RUN systemctl enable snapd.apparmor

# Set the proper environment.
ENV container=docker \
	init=/lib/systemd/systemd

# Add seeding folder for snapd
RUN mkdir -p /var/lib/snapd
COPY seed /var/lib/snapd/seed

# Add store specific network relay script
ADD scripts/network-relay.sh /bin/

# Add our entrypoint
ADD scripts/entrypoint.sh /bin/
ADD scripts/systemd-detect-virt /usr/bin/

# Add scripts
ADD scripts/snap-daemon-starter.sh /usr/local/bin/
ADD scripts/init.sh /usr/local/bin/

# Add service configs 
ADD configs/sshd_config /etc/ssh/
ADD configs/rabbitmq-env.conf /root/

# Add service files
ADD services/init.service /etc/systemd/system/

# Ensure docker sends the shutdown signal that systemd expects
STOPSIGNAL SIGRTMIN+3
	
# Set entrypoint
ENTRYPOINT ["/bin/entrypoint.sh"]

# Set default command
CMD ["snap"]
