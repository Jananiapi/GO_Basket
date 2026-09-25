import java.util.*;
import java.nio.file.*;
import java.io.*;
import com.hazelcast.config.Config;
import com.hazelcast.core.*;
import in.codifi.basket.entity.primary.DeviceMappingEntity;

public class CacheFixture {
 public static void main(String[] args)throws Exception{
  Config c=new Config();c.setClusterName("basket-go-parity");c.getNetworkConfig().setPort(15701).setPortAutoIncrement(false);c.getNetworkConfig().getInterfaces().setEnabled(true).addInterface("127.0.0.1");c.getNetworkConfig().getJoin().getMulticastConfig().setEnabled(false);c.setProperty("hazelcast.logging.type","none");
  HazelcastInstance h=Hazelcast.newHazelcastInstance(c);
  h.getMap("contractMaster").put("NFO_47310",read(args[0]+"/contract.ser"));h.getMap("userKeyMap").put("USER1",read(args[0]+"/customer.ser"));h.getMap("restUserSession").put("USER1_REST_SESSION","session");
  System.out.println("CACHE_READY");System.out.flush();
  long deadline=System.currentTimeMillis()+180000;
  while(System.currentTimeMillis()<deadline){Object value=h.getMap("deviceMappingDetails").get("USER1");if(value!=null){List<DeviceMappingEntity> ds=(List<DeviceMappingEntity>)value;if(ds.size()!=1||!ds.get(0).getUserId().equals("USER1")||ds.get(0).getId()!=7)throw new Exception("invalid Go device payload");System.out.println("JAVA_DEVICE_READ_OK");System.out.flush();break;}Thread.sleep(100);}
  h.shutdown();
 }
 static Object read(String p)throws Exception{try(ObjectInputStream in=new ObjectInputStream(new FileInputStream(p))){return in.readObject();}}
}
