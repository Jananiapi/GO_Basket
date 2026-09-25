import java.io.*;
import java.util.*;
import java.security.*;
import in.codifi.cache.model.ContractMasterModel;
import in.codifi.basket.entity.primary.DeviceMappingEntity;
import com.sas.dto.CustomerDTO;
import com.sas.dto.DefaultLoginDTO;

// Test fixture generator only. The Go service has no JVM dependency.
public class Fixture {
 public static void main(String[] args) throws Exception {
  if (args[0].equals("read-devices")) {
   try(ObjectInputStream in=new ObjectInputStream(new FileInputStream(args[1]))) {
    List<DeviceMappingEntity> ds=(List<DeviceMappingEntity>)in.readObject();
    if(ds.size()!=1 || !ds.get(0).getUserId().equals("USER1") || ds.get(0).getId()!=7 || ds.get(0).getActiveStatus()!=1)throw new Exception("device compatibility failure");
    System.out.println("Go device list deserialized by original Java classes");
   }
   return;
  }
  ContractMasterModel c=new ContractMasterModel(); c.setExch("NFO");c.setToken("47310");c.setTradingSymbol("NIFTY15SEP26P23800");c.setFormattedInsName("NIFTY 15th SEP 23800 PE");c.setLotSize("65");c.setExpiry(new Date(1789410600000L));c.setCompanyName("Test \uD83D\uDE00\u0000");write(args[0]+"/contract.ser",c);
  CustomerDTO customer=new CustomerDTO();customer.setUserId("USER1");customer.setStringPkey4("fixture-key");customer.setTomcatcount("3");DefaultLoginDTO login=new DefaultLoginDTO();login.setS_prdt_ali("CNC:CNC");customer.setUserSettingDto(login);KeyPairGenerator gen=KeyPairGenerator.getInstance("RSA");gen.initialize(512);customer.setPublicKey4(gen.generateKeyPair().getPublic());write(args[0]+"/customer.ser",customer);
 }
 static void write(String path,Object value)throws Exception{try(ObjectOutputStream out=new ObjectOutputStream(new FileOutputStream(path))){out.writeObject(value);}}
}
