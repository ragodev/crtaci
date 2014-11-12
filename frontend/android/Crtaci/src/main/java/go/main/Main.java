// Java Package main is a proxy for talking to a Go program.
//
// File is generated by gobind. Do not edit.
package go.main;

import go.Seq;

public abstract class Main {
    private Main() {} // uninstantiable
    
    public static String Extract(String service, String videoId) {
        go.Seq _in = new go.Seq();
        go.Seq _out = new go.Seq();
        String _result;
        _in.writeUTF16(service);
        _in.writeUTF16(videoId);
        Seq.send(DESCRIPTOR, CALL_Extract, _in, _out);
        _result = _out.readUTF16();
        if(_result == null || _result.isEmpty()) {
            return null;
        }
        return _result;
    }
    
    public static String List() {
        go.Seq _in = new go.Seq();
        go.Seq _out = new go.Seq();
        String _result;
        Seq.send(DESCRIPTOR, CALL_List, _in, _out);
        _result = _out.readUTF16();
        if(_result == null || _result.isEmpty()) {
            return null;
        }
        return _result;
    }
    
    public static void ListenAndServe(String bind) {
        go.Seq _in = new go.Seq();
        go.Seq _out = new go.Seq();
        _in.writeUTF16(bind);
        Seq.send(DESCRIPTOR, CALL_ListenAndServe, _in, _out);
    }
    
    public static String Search(String query) {
        go.Seq _in = new go.Seq();
        go.Seq _out = new go.Seq();
        String _result;
        _in.writeUTF16(query);
        Seq.send(DESCRIPTOR, CALL_Search, _in, _out);
        _result = _out.readUTF16();
        if(_result == null || _result.isEmpty()) {
            return null;
        }
        return _result;
    }
    
    private static final int CALL_Extract = 1;
    private static final int CALL_List = 2;
    private static final int CALL_ListenAndServe = 3;
    private static final int CALL_Search = 4;
    private static final String DESCRIPTOR = "main";
}
