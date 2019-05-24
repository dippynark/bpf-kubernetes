['vagrant-reload', 'vagrant-disksize'].each do |plugin|
  unless Vagrant.has_plugin?(plugin)
    raise "Vagrant plugin #{plugin} is not installed!"
  end
end

Vagrant.configure('2') do |config|
  config.vm.box = "ubuntu/bionic64" # Ubuntu 18.04
  config.disksize.size = "20GB"
  config.vm.network "private_network", type: "dhcp"

  # fix issues with slow dns http://serverfault.com/a/595010
  config.vm.provider :virtualbox do |vb, override|
      vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
      vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
      vb.customize ["modifyvm", :id, "--memory", "2048"]
      vb.customize ["modifyvm", :id, "--cpus", 2]
  end

  config.vm.provision :shell, :privileged => true, :path => "kernel.sh"
  config.vm.provision :reload
  config.vm.provision :shell, :privileged => true, :path => "tools.sh"
  config.vm.provision :shell, :privileged => true, :path => "kubernetes.sh"
end

